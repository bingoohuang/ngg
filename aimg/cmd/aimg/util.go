package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strconv"
	"strings"

	"github.com/cespare/xxhash/v2"
	"gorm.io/gorm"
)

func XxHash(body []byte) string {
	x := xxhash.New()
	_, _ = x.Write(body)
	h := x.Sum64()
	xh := strconv.FormatUint(h, 16)
	return xh
}

func ReadFsFile(fsys fs.FS, path string) ([]byte, error) {
	f, err := fsys.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	body, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", path, err)
	}
	return body, nil
}

// ParsePerceptionHash converts a perception hash string "p:<hex>" to a byte slice.
func ParsePerceptionHash(hashStr string) ([]byte, error) {
	if !strings.HasPrefix(hashStr, "p:") {
		return nil, errors.New("invalid perception hash format: missing 'p:' prefix")
	}

	hexPart := strings.TrimPrefix(hashStr, "p:")
	bytes, err := hex.DecodeString(hexPart)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// HammingDistance calculates the number of differing bits between two byte slices.
func HammingDistance(hash1, hash2 []byte) (int, error) {
	if len(hash1) != len(hash2) {
		return 0, errors.New("hash lengths do not match")
	}

	distance := 0
	for i := 0; i < len(hash1); i++ {
		x := hash1[i] ^ hash2[i]
		for x != 0 {
			distance++
			x &= x - 1
		}
	}

	return distance, nil
}

var hammingThreshold = 10

// HammingGroup represents a hash -> group assignment used for
// inmemory hammingdistance calulations.
type HammingGroup struct {
	GroupID uint
	Hash    []byte
}

var hammingGroups []HammingGroup

// AssignGroupID assigns a PerceptionHashGroupId based on Hamming distance
func AssignGroupID(db *gorm.DB, perceptionHashStr string) (uint, error) {
	// Parse the incoming perception hash
	parsedHash, err := ParsePerceptionHash(perceptionHashStr)
	if err != nil {
		return 0, err
	}

	// Iterate through existing groups to find a match
	for _, group := range hammingGroups {
		dist, err := HammingDistance(parsedHash, group.Hash)
		if err != nil {
			return 0, err
		}

		if dist <= hammingThreshold {
			return group.GroupID, nil
		}
	}

	// No matching group found; create a new group
	var maxGroupID uint
	err = db.Model(&Img{}).
		Select("COALESCE(MAX(perception_hash_group_id), 0)").
		Scan(&maxGroupID).Error
	if err != nil {
		return 0, err
	}
	nextGroupID := maxGroupID + 1

	// Add the new group to in-memory cache
	newGroup := HammingGroup{
		GroupID: nextGroupID,
		Hash:    parsedHash,
	}
	hammingGroups = append(hammingGroups, newGroup)

	return nextGroupID, nil
}
