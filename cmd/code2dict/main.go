package main

import (
	"code2dict/pkg/cmdutils"
	"code2dict/pkg/filestats"
	"code2dict/pkg/generate"
	"code2dict/pkg/logging"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jessevdk/go-flags"
)

func main() {
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	// 添加描述信息
	parser.Name = AppName
	parser.Usage = "[OPTIONS]"
	parser.ShortDescription = AppShortDesc
	parser.LongDescription = AppLongDesc

	if _, err := parser.Parse(); err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && errors.Is(flagsErr.Type, flags.ErrHelp) {
			return
		}
		fmt.Printf("命令行参数解析错误: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logCfg := logging.NewLogConfig(opts.LogLevel, opts.LogFile, opts.ConsoleFormat)
	if err := logging.InitLogger(logCfg); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logging.Sync()

	// 新增：判断是否需要显示版本信息
	if opts.Version {
		fmt.Printf("CodeClear version %s\n", AppVersion)
		fmt.Printf("Build Date: %s\n", BuildDate)
		os.Exit(0) // 显示后退出，不执行后续逻辑
	}

	// 检查是否输入 Path
	if opts.Path == "" {
		logging.Fatalf("必须有指定代码文件所在目录!!!")
	}

	// 统计模式 目录大小统计
	if opts.StatsDir {
		if err := filestats.RunStatsDir(opts.Path); err != nil {
			logging.Fatalf("目录统计操作失败: %v", err)
		}
		os.Exit(0) // 显示后退出，不执行后续逻辑
	}

	// 统计模式 文件类型统计
	if opts.StatsExt {
		if err := filestats.RunStatsExt(opts.Path); err != nil {
			logging.Fatalf("后缀统计操作失败: %v", err)
		}
		os.Exit(0) // 显示后退出，不执行后续逻辑
	}

	// 自动生成输出文件名
	outputFile := opts.Output
	if outputFile == "" {
		// 根据输入路径自动生成输出文件名
		baseName := filepath.Base(filepath.Clean(opts.Path))
		outputFile = fmt.Sprintf("%s.dict.txt", baseName)
	}

	// 按后缀进行字典生成
	if opts.Preset != "" {
		preset := initPresetConfig(opts.Preset, opts.PresetConfig)

		// 创建字典生成器并运行
		if preset != nil {
			if (opts.EnWhite && len(preset.Include) > 0) || (!opts.EnWhite && len(preset.Include)+len(preset.Exclude)+len(preset.Ignored) > 0) {
				dictGenerator := generate.NewDictGenerator(opts.Path, *preset, opts.EnWhite, opts.EnCover, outputFile)
				if err := dictGenerator.RunGenerate(); err != nil {
					logging.Fatalf("生成文件路径字典失败: %v", err)
				}
			} else {
				logging.Fatalf("当前 Preset (%s) 未配置有效数据: %s", opts.Preset, cmdutils.AnyToJson(preset))
			}
		} else {
			logging.Fatalf("当前 Preset (%s) 配置初始化详细配置失败!", opts.Preset)
		}
	}
}
