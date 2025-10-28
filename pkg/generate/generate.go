package generate

import (
	"code2dict/internal/config"
	"code2dict/pkg/fileutils"
	"code2dict/pkg/logging"
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

// DictGenerator 字典生成器选项
type DictGenerator struct {
	Path    string   // 根目录路径
	Allowed []string // 白名单扩展名列表
	Removed []string // 移除扩展名列表（兼容旧参数）
	Ignored []string // 待删除目录名列表（仅匹配目录名，不限制层级）
	EnWhite bool     // 白名单模式（保留Allowed列表，删除其他）
	EnCover bool     // 覆盖写入模式，默认为追加写入
	Output  string   // 输出文件路径
}

// NewDictGenerator 创建字典生成器实例
func NewDictGenerator(path string, preset config.PresetConfig, enWhite, enCover bool, output string) *DictGenerator {
	return &DictGenerator{
		Path:    path,
		Allowed: preset.Allowed,
		Removed: preset.Removed,
		Ignored: preset.Ignored,
		EnWhite: enWhite,
		EnCover: enCover,
		Output:  output,
	}
}

// RunGenerate 执行字典生成操作
func (g *DictGenerator) RunGenerate() error {

	// 提前检查配置是否有效，避免进行无效操作
	if g.EnWhite {
		// 白名单模式处理 （使用allowed列表）
		if len(g.Allowed) == 0 {
			return fmt.Errorf("白名单模式未配置allowed列表，无法继续处理")
		}
	} else {
		// 普通模式处理（使用Removed列表）
		if len(g.Removed) == 0 && len(g.Ignored) == 0 {
			return fmt.Errorf("黑名单模式未配置 removed 列表或 ignored 列表，无法继续处理")
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
	allowed := preprocessExtensions(g.Allowed)
	removed := preprocessExtensions(g.Removed)
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
			g.handleFiles(path, allowed, removed, rootAbsPath, &paths)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("字典生成过程中发生错误: %v", err)
	}

	// 将所有路径写入文件
	if err := fileutils.WritePathsToFile(g.Output, paths, g.EnCover); err != nil {
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
func (g *DictGenerator) handleFiles(path string, allowedExts, removedExts []string, rootPath string, paths *[]string) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))

	if g.EnWhite {
		if isExtensionInList(ext, allowedExts) {
			g.generatePath(path, rootPath, paths)
		}
	} else {
		if !isExtensionInList(ext, removedExts) {
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
