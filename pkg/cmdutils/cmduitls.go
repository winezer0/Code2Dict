package cmdutils

import (
	"encoding/json"
	"regexp"
	"strings"
)

func ParseCmdExtDir(input string) (extList, dirList []string) {
	extList = []string{}
	dirList = []string{}

	// 兼容低版本 Go 的正则：不使用零宽断言，而是匹配到下一个标记或结尾
	// 模式说明：
	// (ext|dir):   匹配 ext: 或 dir:
	// (.*?)        非贪婪匹配内容（直到下一个标记或结尾）
	// (?:ext:|dir:|$)  匹配下一个标记（非捕获组）或字符串结尾
	re := regexp.MustCompile(`(ext|dir):(.*?)(ext:|dir:|$)`)

	remaining := input // 剩余未处理的字符串
	for {
		// 查找匹配
		match := re.FindStringSubmatch(remaining)
		if len(match) != 4 {
			break // 无更多匹配，退出循环
		}

		key := strings.ToLower(match[1])
		value := strings.TrimSpace(match[2])
		nextMarker := match[3] // 下一个标记（可能为空，即到结尾）

		// 处理当前值
		items := strings.Split(value, ",")
		for _, item := range items {
			item = strings.TrimSpace(item)
			if item != "" {
				switch key {
				case "ext":
					extList = append(extList, item)
				case "dir":
					dirList = append(dirList, item)
				}
			}
		}

		// 移动到下一个标记的位置继续处理
		remaining = remaining[len(match[1])+1+len(match[2]):] // +1 是因为 key 后面有个冒号
		if nextMarker == "" {
			break // 已到结尾，退出
		}
	}

	return extList, dirList
}

// ListUnique 去除字符串切片中的重复元素
// 如果 tolower 为 true，则忽略大小写进行去重，并返回小写字符串
// 如果 tolower 为 false，则区分大小写去重，返回原始字符串
func ListUnique(input []string, tolower bool) []string {
	seen := make(map[string]struct{}, len(input))
	result := make([]string, 0, len(input)) // 预分配容量

	for _, s := range input {
		key := s
		if tolower {
			key = strings.ToLower(s)
		}

		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			result = append(result, key)
		}
	}

	return result
}

func AnyToJson(v any) string {
	if bytes, err := json.Marshal(v); err != nil {
		return ""
	} else {
		return string(bytes)
	}
}
