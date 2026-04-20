package main

import (
	"code2dict/internal/config"

	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/winezer0/xutils/logging"
	"github.com/winezer0/xutils/utils"
)

// 版本信息常量（根据实际情况修改）
const (
	AppName      = "Code2Dict"
	AppShortDesc = "ISEC 代码文件路径字典生成工具"
	AppLongDesc  = "ISEC 代码文件路径字典生成工具, 生成指定目录中文件的URL路径字典"
	AppVersion   = "0.0.4"
	BuildDate    = "2026-04-20"
)

// Options command line options
type Options struct {
	SourcePath string `short:"p" long:"source" description:"扫描起始目录路径"`
	PresetName string `short:"P" long:"preset" description:"使用预设规则 或 ext/dir:逗号分割的后缀列表 (如ext: exe,txt)" default:"common"`
	ConfigPath string `short:"c" long:"config" description:"自定义 YAML 配置文件路径"`
	OutputFile string `short:"o" long:"output" description:"输出字典文件路径"`

	WhiteListMode bool `short:"w" long:"white" description:"白名单模式：仅保留预设中 include 指定的文件后缀类型"`
	OverWriteMode bool `short:"W" long:"cover" description:"使用覆盖写入模式到结果文件"`

	// GenerateConfig 生成默认配置文件
	GenerateConfig bool `long:"gen" description:"生成默认配置文件到<ConfigPath>"`
	ShowPresetList bool `long:"list" description:"列出<ConfigPath>配置文件中的所有预设值"`

	// 统计信息显示
	StatsExt bool `short:"s" long:"stats_ext" description:"启用统计模式：显示目录下(后缀类型) 数量分布"`
	StatsDir bool `short:"S" long:"stats_dir" description:"启用统计模式：显示目录下(目录文件) 数量分布"`
	Version  bool `short:"v" long:"version" description:"输出版本信息"`

	// Log configuration
	LogFile       string `long:"lf" description:"Log file path (default: null)"`
	LogLevel      string `long:"ll" description:"Log level (debug/info/warn/error)" default:"info"`
	ConsoleFormat string `long:"cf" description:"Console log format (T L C M F combination or off|null to disable)" default:"M"`
}

// InitOptionsArgs 常用的工具函数，解析parser和logging配置
func InitOptionsArgs(minimumParams int) (*Options, *flags.Parser) {
	opts := &Options{}
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = AppName
	parser.Usage = "[OPTIONS]"
	parser.ShortDescription = AppShortDesc
	parser.LongDescription = AppLongDesc

	// 命令行参数数量检查 指不包含程序名本身的参数数量
	if minimumParams > 0 && len(os.Args)-1 < minimumParams {
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	// 命令行参数解析检查
	if _, err := parser.Parse(); err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && errors.Is(flagsErr.Type, flags.ErrHelp) {
			os.Exit(0)
		}
		fmt.Printf("Error:%v\n", err)
		os.Exit(1)
	}

	// 版本号输出
	if opts.Version {
		fmt.Printf("%s version %s\n", AppName, AppVersion)
		fmt.Printf("Build Date: %s\n", BuildDate)
		os.Exit(0)
	}

	// 初始化日志器
	logCfg := logging.NewLogConfig(opts.LogLevel, opts.LogFile, opts.ConsoleFormat)
	if err := logging.InitLogger(logCfg); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// 处理生成配置文件命令
	if opts.GenerateConfig {
		configPath := opts.ConfigPath
		if configPath == "" {
			configPath = AppName + ".yaml"
		}
		if err := config.GenDefaultConfig(configPath); err != nil {
			logging.Fatalf("Failed to generate config file: %v", err)
		}
		logging.Infof("Default config file has been generated: %s", configPath)
		os.Exit(0)
	}

	// 输出当前配置文件中的所有preset名称
	if opts.ShowPresetList {
		conf, err := config.LoadConfig(opts.ConfigPath, AppName)
		if err != nil {
			logging.Fatalf("load config: %s error: %v", conf, err)
		}
		config.PrintPresetSummary(conf)
		os.Exit(0)
	}
	// 检查是否输入 Path
	if opts.SourcePath == "" {
		logging.Fatalf("必须有指定代码文件所在目录!!!")
	}

	// 自动生成输出文件名
	if opts.OutputFile == "" {
		// 根据输入路径自动生成输出文件名
		baseName := filepath.Base(filepath.Clean(opts.SourcePath))
		opts.OutputFile = fmt.Sprintf("%s.dict.txt", baseName)
	}

	return opts, parser
}

func initPresetConfig(presetStr string, presetFile string) *config.PresetConfig {
	// 获取preset配置
	var preset *config.PresetConfig
	if strings.Contains(presetStr, "ext:") || strings.Contains(presetStr, "dir:") {
		// 从输入命令行中解析出 preset
		extList, dirList := parseCmdExtDir(presetStr)
		extList = utils.UniqueSlice(utils.ToLowerKeys(extList), true, true)
		dirList = utils.UniqueSlice(utils.ToLowerKeys(dirList), true, true) // 仅在黑名单模式下有效,用于删除自定义目录，很少用
		preset = config.NewPresetConfig("temp list", extList, extList, dirList)
		logging.Infof("cmd init preset: %s", utils.ToJSON(preset))
	} else {
		conf, err := config.LoadConfig(presetFile, AppName)
		if err != nil {
			logging.Errorf("load config: %s error: %v", conf, err)
		} else {
			if preset, _ = conf.GetPreset(presetStr); preset == nil {
				logging.Errorf("config %s not contain key: %s and custom preset not like (like ext:xxx,xxx)", conf, presetStr)
			}
		}
	}
	return preset
}

// parseCmdExtDir 解析和格式化命令行參數中的dir和ext参数
func parseCmdExtDir(input string) (extList, dirList []string) {
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
