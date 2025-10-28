package filestats

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// DirStats 目录统计信息
type DirStats struct {
	Path  string
	Count int
	Size  int64
}

// RunStatsDir 执行目录统计功能
func RunStatsDir(path string) error {
	// 收集目录统计信息
	stats, err := collectDirsStats(path)
	if err != nil {
		return err
	}

	// 排序目录统计信息
	sortedStats := sortDirsStats(stats)

	// 输出目录统计信息
	printDirsStats(sortedStats, path)

	return nil
}

// collectDirsStats 收集目录统计信息
func collectDirsStats(rootPath string) (map[string]*DirStats, error) {
	stats := make(map[string]*DirStats)

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// 忽略无法访问的文件或目录，继续处理其他文件
			return nil
		}

		if info.IsDir() {
			// 初始化目录统计信息
			if _, exists := stats[path]; !exists {
				stats[path] = &DirStats{
					Path: path,
				}
			}
			return nil
		}

		// 获取文件所在目录
		dir := filepath.Dir(path)

		// 更新目录统计信息
		if stat, exists := stats[dir]; exists {
			stat.Count++
			stat.Size += info.Size()
		} else {
			stats[dir] = &DirStats{
				Path:  dir,
				Count: 1,
				Size:  info.Size(),
			}
		}

		// 同时更新所有父目录的统计信息
		parentDir := dir
		for parentDir != rootPath && parentDir != "." && parentDir != "/" {
			parentDir = filepath.Dir(parentDir)
			if stat, exists := stats[parentDir]; exists {
				stat.Count++
				stat.Size += info.Size()
			} else {
				stats[parentDir] = &DirStats{
					Path:  parentDir,
					Count: 1,
					Size:  info.Size(),
				}
			}
		}

		return nil
	})

	return stats, err
}

// sortDirsStats 排序目录统计信息
func sortDirsStats(stats map[string]*DirStats) []*DirStats {
	// 转换为切片
	var statsList []*DirStats
	for _, stat := range stats {
		statsList = append(statsList, stat)
	}

	// 按大小降序排序
	sort.Slice(statsList, func(i, j int) bool {
		if statsList[i].Size != statsList[j].Size {
			return statsList[i].Size > statsList[j].Size
		}
		// 如果大小相同，按文件数量排序
		return statsList[i].Count > statsList[j].Count
	})

	return statsList
}

// printDirsStats 输出目录统计信息
func printDirsStats(statsList []*DirStats, rootPath string) {
	// 输出统计结果
	fmt.Printf("\n目录文件大小统计 (%s):\n", rootPath)
	fmt.Printf("%-100s %-10s %-15s\n", "Directory", "Files", "Size")
	fmt.Printf("%-100s %-10s %-15s\n", "----------", "-----", "----")

	for _, stat := range statsList {
		// 只显示包含文件的目录
		if stat.Count > 0 {
			fmt.Printf("%-100s %-10d %-15s\n", stat.Path, stat.Count, formatSize(stat.Size))
		}
	}

	fmt.Println()
}

// formatSize 将字节大小格式化为合适的单位
func formatSize(size int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	unitIndex := 0
	floatSize := float64(size)

	for floatSize >= 1024 && unitIndex < len(units)-1 {
		floatSize /= 1024
		unitIndex++
	}

	if unitIndex == 0 {
		return fmt.Sprintf("%d %s", size, units[unitIndex])
	}

	return fmt.Sprintf("%.1f %s", floatSize, units[unitIndex])
}
