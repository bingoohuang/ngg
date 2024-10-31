package ggtrand

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

var charset = []byte(
	"abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" +
		"!@#$%^&*()_+-=[]{}|;:',.<>?/`~\"\\ ")

// GenerateRandomTextFile 生成一个指定长度的随机文本文件，返回文件名
func GenerateRandomTextFile(length int) (string, error) {
	if length < 1 {
		return "", fmt.Errorf("length must be greater than 0")
	}

	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "*")
	if err != nil {
		return "", err
	}
	// 创建临时文件
	tmpFile, err := os.CreateTemp(tmpDir, "*.txt")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	// 初始化随机数生成器
	rand.Seed(time.Now().UnixNano())

	// 生成随机文本并写入文件
	currentLineLength := 0
	lineLength := 120

	for i := 0; i < length; i++ {
		if currentLineLength == lineLength-1 { // 到达行末，添加换行符
			tmpFile.WriteString("\n")
			currentLineLength = 0
		} else {
			j := rand.Intn(len(charset))
			tmpFile.Write(charset[j : j+1])
			currentLineLength++
		}
	}

	return tmpFile.Name(), nil
}
