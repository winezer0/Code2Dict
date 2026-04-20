package main

import (
	"code2dict/pkg/filestats"
	"code2dict/pkg/generate"

	"os"

	"github.com/winezer0/xutils/logging"
	"github.com/winezer0/xutils/utils"
)

func main() {
	// 打印命令行输入配置
	opts, _ := InitOptionsArgs(1)
	defer logging.Sync()

	// 统计模式 目录大小统计
	if opts.StatsDir {
		if err := filestats.RunStatsDir(opts.SourcePath); err != nil {
			logging.Fatalf("目录统计操作失败: %v", err)
		}
		os.Exit(0) // 显示后退出，不执行后续逻辑
	}

	// 统计模式 文件类型统计
	if opts.StatsExt {
		if err := filestats.RunStatsExt(opts.SourcePath); err != nil {
			logging.Fatalf("后缀统计操作失败: %v", err)
		}
		os.Exit(0) // 显示后退出，不执行后续逻辑
	}

	// 按后缀进行字典生成
	if opts.PresetName != "" {
		preset := initPresetConfig(opts.PresetName, opts.ConfigPath)

		// 创建字典生成器并运行
		if preset != nil {
			if (opts.WhiteListMode && len(preset.Include) > 0) || (!opts.WhiteListMode && len(preset.Include)+len(preset.Exclude)+len(preset.Ignored) > 0) {
				dictGenerator := generate.NewDictGenerator(opts.SourcePath, opts.OutputFile, *preset, opts.WhiteListMode, opts.OverWriteMode)
				if err := dictGenerator.RunGenerate(); err != nil {
					logging.Fatalf("生成文件路径字典失败: %v", err)
				}
			} else {
				logging.Fatalf("当前 Preset (%s) 未配置有效数据: %s", opts.PresetName, utils.ToJSON(preset))
			}
		} else {
			logging.Fatalf("当前 Preset (%s) 配置初始化详细配置失败!", opts.PresetName)
		}
	}
}
