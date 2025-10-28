package main

import (
	"code2dict/internal/config"
	"code2dict/internal/embeds"
	"code2dict/pkg/cmdutils"
	"code2dict/pkg/filestats"
	"code2dict/pkg/fileutils"
	"code2dict/pkg/generate"
	"code2dict/pkg/logging"
	"errors"
	"fmt"
	"github.com/jessevdk/go-flags"
	"os"
	"path/filepath"
	"strings"
)

// 版本信息常量（根据实际情况修改）
const (
	AppName      = "Code2Dict"
	AppShortDesc = "ISEC 代码文件路径字典生成工具"
	AppLongDesc  = "ISEC 代码文件路径字典生成工具, 生成指定目录中文件的URL路径字典"
	AppVersion   = "0.0.2"
	BuildDate    = "2025-10-28"
)

// Options command line options
type Options struct {
	Path         string `short:"p" long:"path" description:"扫描起始目录路径" required:"true"`
	Preset       string `short:"P" long:"preset" description:"使用预设规则(默认common) 或 ext/dir:逗号分割的后缀列表 (如ext: exe,txt)"`
	PresetConfig string `short:"c" long:"preset_config" description:"自定义 YAML 配置文件路径" default:"code2dict.yaml"`
	Output       string `short:"o" long:"output" description:"输出字典文件路径"`

	EnWhite bool `short:"w" long:"en_white" description:"白名单模式：仅保留预设中 include 指定的文件后缀类型"`
	EnCover bool `short:"W" long:"en_cover" description:"使用覆盖写入模式到结果文件"`

	// 统计信息显示
	StatsExt bool `short:"s" long:"stats_ext" description:"启用统计模式：显示目录下(后缀类型) 数量分布"`
	StatsDir bool `short:"S" long:"stats_dir" description:"启用统计模式：显示目录下(目录文件) 数量分布"`
	Version  bool `short:"v" long:"version" description:"输出版本信息"`

	// Log configuration
	LogFile       string `long:"lf" description:"Log file path (default: null)"`
	LogLevel      string `long:"ll" description:"Log level (debug/info/warn/error)" default:"info"`
	ConsoleFormat string `long:"cf" description:"Console log format (T L C M F combination or off|null to disable)" default:"M"`
}

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

func initPresetConfig(presetStr string, presetFile string) *config.PresetConfig {
	// 获取preset配置
	var preset *config.PresetConfig
	if strings.Contains(presetStr, "ext:") || strings.Contains(presetStr, "dir:") {
		// 从输入命令行中解析出 preset
		extList, dirList := cmdutils.ParseCmdExtDir(presetStr)
		extList = cmdutils.ListUnique(extList, true)
		dirList = cmdutils.ListUnique(dirList, true) // 仅在黑名单模式下有效,用于删除自定义目录，很少用
		preset = config.NewPresetConfig("临时名单", extList, extList, dirList)
		logging.Infof("cmd init preset: %s", cmdutils.AnyToJson(preset))
	} else {
		// 从配置文件中获取 preset
		checkAndInitPresetFile(presetFile)
		if conf, err := config.LoadConfig(presetFile); err != nil {
			logging.Errorf("load config: %s error: %v", conf, err)
		} else if preset, _ = conf.GetPreset(presetStr); preset == nil {
			logging.Errorf("config %s not contain key: %s and custom preset not like (like ext:xxx,xxx)", conf, presetStr)
		}
	}
	return preset
}

// checkAndInitPresetFile presetFile为空时生成默认配置
func checkAndInitPresetFile(presetFile string) {
	if fileutils.IsEmptyFile(presetFile) {
		fileutils.MakeDirs(presetFile, true)
		fileutils.WriteAny(presetFile, embeds.GetConfig())
		logging.Debugf("Success creat config from embed: %v", presetFile)
	}
}
