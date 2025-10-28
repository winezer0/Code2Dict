package filestats

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileStats 文件统计信息
type FileStats struct {
	Extension string
	Count     int
}

// RunStatsExt 执行统计功能
func RunStatsExt(path string) error {
	// 执行统计
	statsMap, err := collectStats(path)
	if err != nil {
		return err
	}

	// 执行排序
	statsList := sortStats(statsMap)

	// 执行输出
	printStats(statsList, path)

	return nil
}

// collectStats 进行数量统计
func collectStats(path string) (map[string]*FileStats, error) {
	statsMap := make(map[string]*FileStats)

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// 忽略无法访问的文件或目录，继续处理其他文件
			return nil
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
		if ext == "" {
			ext = "NONE" // 使用NONE表示无后缀文件
		} else {
			ext = "." + ext
		}

		if stat, exists := statsMap[ext]; exists {
			stat.Count++
		} else {
			statsMap[ext] = &FileStats{
				Extension: ext,
				Count:     1,
			}
		}

		return nil
	})

	return statsMap, err
}

// sortStats 进行后缀排序
func sortStats(statsMap map[string]*FileStats) []*FileStats {
	// 转换为切片
	var statsList []*FileStats
	for _, stat := range statsMap {
		statsList = append(statsList, stat)
	}

	// 排序
	sort.Slice(statsList, func(i, j int) bool {
		if statsList[i].Count != statsList[j].Count {
			return statsList[i].Count > statsList[j].Count
		}
		// 如果计数相同，按扩展名排序
		return statsList[i].Extension < statsList[j].Extension
	})

	return statsList
}

// printStats 进行结果输出
func printStats(statsList []*FileStats, path string) {
	// 输出统计结果
	fmt.Printf("\n文件类型统计 (%s):\n", path)
	fmt.Printf("%-20s %-10s\n", "Suffix", "Counts")
	fmt.Printf("%-20s %-10s\n", "------", "------")

	// 提取所有后缀，构建列表
	var extensions []string
	for _, stat := range statsList {
		extensions = append(extensions, stat.Extension)
		fmt.Printf("%-20s %-10d\n", stat.Extension, stat.Count)
	}
	fmt.Printf("\n所有后缀 (%d): %s\n", len(extensions), strings.Join(extensions, ","))
}
