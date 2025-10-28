package generate

import (
	"strings"
)

// 检查扩展名是否在目标列表中
func isExtensionInList(ext string, extensions []string) bool {
	for _, e := range extensions {
		if e == "none" && ext == "" {
			return true
		}
		if ext == e {
			return true
		}
	}
	return false
}

// 检查Dir是否在目标列表中
func isDirInList(dir string, dirs []string) bool {
	for _, e := range dirs {
		if dir == e {
			return true
		}
	}
	return false
}

// 预处理扩展名（统一转为小写，去除前缀点）
func preprocessExtensions(exts []string) []string {
	result := make([]string, 0, len(exts))
	for _, e := range exts {
		cleaned := strings.ToLower(strings.TrimPrefix(e, "."))
		result = append(result, cleaned)
	}
	return result
}

// 预处理目录名（统一格式，可选大小写转换）
func preprocessDirNames(names []string) []string {
	result := make([]string, 0, len(names))
	for _, name := range names {
		cleaned := strings.TrimSpace(strings.ToLower(name))
		if cleaned != "" {
			result = append(result, cleaned)
		}
	}
	return result
}
