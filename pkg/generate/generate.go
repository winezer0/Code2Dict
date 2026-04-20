package generate

import (
	"bufio"
	"code2dict/internal/config"

	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/winezer0/xutils/logging"
)

// DictGenerator 字典生成器选项
type DictGenerator struct {
	Path      string   // 根目录路径
	Include   []string // 白名单扩展名列表
	Exclude   []string // 移除扩展名列表（兼容旧参数）
	Ignored   []string // 待删除目录名列表（仅匹配目录名，不限制层级）
	WhiteMode bool     // 白名单模式（保留Include列表，删除其他）
	OverMode  bool     // 覆盖写入模式，默认为追加写入
	Output    string   // 输出文件路径
}

// NewDictGenerator 创建字典生成器实例
func NewDictGenerator(path, output string, preset config.PresetConfig, whiteMode, OverMode bool) *DictGenerator {
	return &DictGenerator{
		Path:      path,
		Output:    output,
		Include:   preset.Include,
		Exclude:   preset.Exclude,
		Ignored:   preset.Ignored,
		WhiteMode: whiteMode,
		OverMode:  OverMode,
	}
}

// RunGenerate 执行字典生成操作
func (g *DictGenerator) RunGenerate() error {

	// 提前检查配置是否有效，避免进行无效操作
	if g.WhiteMode {
		// 白名单模式处理 （使用include列表）
		if len(g.Include) == 0 {
			return fmt.Errorf("白名单模式未配置include列表，无法继续处理")
		}
	} else {
		// 普通模式处理（使用Exclude列表）
		if len(g.Exclude) == 0 && len(g.Ignored) == 0 {
			return fmt.Errorf("黑名单模式未配置 exclude 列表或 ignored 列表，无法继续处理")
		}
	}

	// 上下文管理（处理中断信号）
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-sigChan:
			logging.Warnf("\n收到中断信号，正在停止操作...")
			cancel()
		case <-ctx.Done():
		}
	}()

	// 预处理配置参数（统一格式）
	include := preprocessExtensions(g.Include)
	exclude := preprocessExtensions(g.Exclude)
	ignored := preprocessDirNames(g.Ignored)

	// 验证根路径有效性
	rootAbsPath, err := filepath.Abs(g.Path)
	if err != nil {
		return fmt.Errorf("目录路径无效: %v", err)
	}

	// 存储生成的路径
	var paths []string

	// 替换为 filepath.WalkDir 提升性能
	err = filepath.WalkDir(rootAbsPath, func(path string, d os.DirEntry, err error) error {
		// 检查中断信号
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 处理文件访问错误
		if err != nil {
			logging.Warnf("访问路径失败: %s - %v", path, err)
			return nil
		}

		// 获取文件信息（WalkDir需要显式获取，减少不必要的系统调用）
		info, err := d.Info()
		if err != nil {
			logging.Warnf("获取信息失败: %s - %v", path, err)
			return nil
		}

		// 检查是否应该跳过当前目录
		if info.IsDir() && len(ignored) > 0 {
			if g.shouldSkipDir(path, ignored) {
				return filepath.SkipDir
			}
		}

		if !info.IsDir() {
			// 非目录文件处理
			g.handleFiles(path, include, exclude, rootAbsPath, &paths)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("字典生成过程中发生错误: %v", err)
	}

	// 将所有路径写入文件
	if err := WritePathsToFile(g.Output, paths, g.OverMode); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	logging.Infof("\nDict Generation completed! Path %s -> Counts: %d", g.Output, len(paths))
	return nil
}

// shouldSkipDir 检查是否应该跳过目录（当目录名匹配 Ignored dirs时）
func (g *DictGenerator) shouldSkipDir(path string, ignored []string) bool {
	currDirName := strings.ToLower(filepath.Base(path))

	// 检查当前目录名是否在目标列表中
	if isDirInList(currDirName, ignored) {
		logging.Infof("跳过目录及其子目录和文件: %s", path)
		return true
	}
	return false
}

// 处理文件路径生成逻辑
func (g *DictGenerator) handleFiles(path string, includeExts, excludeExts []string, rootPath string, paths *[]string) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))

	if g.WhiteMode {
		if isExtensionInList(ext, includeExts) {
			g.generatePath(path, rootPath, paths)
		}
	} else {
		if !isExtensionInList(ext, excludeExts) {
			g.generatePath(path, rootPath, paths)
		}
	}
}

// 生成文件路径并添加到路径列表中
func (g *DictGenerator) generatePath(path string, rootPath string, paths *[]string) {
	// 获取相对路径
	relPath, err := filepath.Rel(rootPath, path)
	if err != nil {
		logging.Warnf("计算相对路径失败: %s - %v", path, err)
		return
	}

	// 转换为URL路径格式
	urlPath := "/" + strings.ReplaceAll(relPath, "\\", "/")

	// 添加到路径列表
	*paths = append(*paths, urlPath)
}

// WritePathsToFile 将路径列表写入文件，mode 以 'a' 开头表示追加，其余为覆盖
func WritePathsToFile(filepath string, paths []string, cover bool) error {
	// 确定打开模式
	flag := os.O_CREATE | os.O_WRONLY
	if cover {
		flag |= os.O_TRUNC // 覆盖写入
	} else {
		flag |= os.O_APPEND // 追加写入
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
