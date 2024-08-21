package sqliter

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DbFile 数据库文件对象
type DbFile struct {
	// Table 表名, e.g. disk
	Table string
	// DividedBy 时间划分, e.g. month.202407
	DividedBy string

	// 主数据库文件, e.g. testdata/metric.t.disk.month.202407.db
	File File

	// Relatives 关联的文件，主要用于计算数据空间大小
	// e.g. testdata/metric.t.disk.month.202407.db 关联
	//     testdata/metric.t.disk.month.202407.db-shm
	//     testdata/metric.t.disk.month.202407.db-wal
	Relatives []File
}

func (f *DbFile) TotalSize() int64 {
	total := f.File.Size
	for _, f := range f.Relatives {
		total += f.Size
	}
	return total
}

type File struct {
	Path string
	Size int64 // 文件名
}

// ListDiskTables 列出磁盘上的所有数据库文件
func (q *Sqliter) ListDiskTables() (map[string][]*DbFile, error) {
	result := map[string][]*DbFile{}
	base := filepath.Base(q.Prefix)
	dirPath := filepath.Dir(q.Prefix)

	fileMap := map[string]*DbFile{} // path => *DbFile
	var leftRelatives []File
	dividedPrefix := q.DividedPrefix()
	if err := filepath.WalkDir(dirPath, func(path string, info os.DirEntry, err error) error {
		name := info.Name()
		if !info.IsDir() && strings.HasPrefix(name, base) {
			if strings.HasSuffix(name, ".db") {
				name = name[:len(name)-3]
				if p := strings.Index(name, dividedPrefix); p > 0 {
					dividedBy := name[p:]
					table := name[len(base) : p-1]

					fileInfo := &DbFile{
						File:      File{Path: path},
						Table:     table,
						DividedBy: dividedBy,
					}
					if fi, _ := info.Info(); fi != nil {
						fileInfo.File.Size = fi.Size()
					}
					fileMap[path] = fileInfo
					result[table] = append(result[table], fileInfo)
				}
			} else if strings.Contains(name, ".db") {
				rel := File{Path: path}
				if fi, _ := info.Info(); fi != nil {
					rel.Size = fi.Size()
				}
				leftRelatives = append(leftRelatives, rel)
			}
		}
		return err
	}); err != nil {
		return nil, err
	}

	for _, rel := range leftRelatives {
		if found := findMainFile(fileMap, rel); !found {
			log.Printf("W! relative file %s is unconnected", rel.Path)
		}
	}

	for k, v := range result {
		sort.Slice(v, func(i, j int) bool {
			return v[i].DividedBy < v[j].DividedBy
		})
		result[k] = v
	}

	return result, nil
}

func findMainFile(fileMap map[string]*DbFile, rel File) bool {
	for k, v := range fileMap {
		if strings.HasPrefix(rel.Path, k) {
			v.Relatives = append(v.Relatives, rel)
			return true
		}
	}
	return false
}
