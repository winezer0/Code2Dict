package fileutils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MakeDirs 创建指定路径的所有必要目录（递归创建）
func MakeDirs(path string, isFile bool) error {
	if isFile {
		path, _ = GetFileDirectory(path)
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create the directory: %w", err)
	}
	return nil
}

// GetFileDirectory 返回给定文件路径的目录部分
func GetFileDirectory(filePath string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("文件路径不能为空")
	}

	dir := filepath.Dir(filePath)
	return dir, nil
}

// IsEmptyFile 检查文件是否为空或不存在
func IsEmptyFile(filename string) bool {
	// Get file info
	fileInfo, err := os.Stat(filename)
	if os.IsNotExist(err) || fileInfo.Size() == 0 {
		return true
	}
	return false
}

// WriteAny 将任意数据写入文本文件
func WriteAny(filePath string, data interface{}) error {
	// 将任意数据转换为字符串形式
	content := fmt.Sprintf("%+v", data)

	// 写入文件
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("file write failed: %w", err)
	}

	return nil
}

// WritePathsToFile 将路径列表写入文件，mode 以 'a' 开头表示追加，其余为覆盖
func WritePathsToFile(filepath string, paths []string, mode string) error {
	// 确定打开模式
	flag := os.O_CREATE | os.O_WRONLY
	if strings.HasPrefix(mode, "a") {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC // "w" 或默认行为：覆盖
	}

	if len(paths) == 0 {
		return fmt.Errorf("paths is empty")
	}

	// 打开文件
	file, err := os.OpenFile(filepath, flag, 0644)
	if err != nil {
		return fmt.Errorf("failed to open the file: %v", err)
	}
	defer file.Close()

	// 使用缓冲写入提升性能（简洁引入 bufio）
	writer := bufio.NewWriter(file)
	for _, path := range paths {
		if _, err := writer.WriteString(strings.TrimSpace(path) + "\n"); err != nil {
			return fmt.Errorf("failed to write to path: %s - %v", path, err)
		}
	}

	// 刷新缓冲区
	return writer.Flush()
}
