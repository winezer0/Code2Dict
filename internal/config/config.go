package config

import (
	"code2dict/internal/embeds"
	"fmt"

	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config 配置文件结构
type Config struct {
	Presets map[string]PresetConfig `yaml:"presets"`
}

// LoadConfig 加载配置文件
// 当用户没有指定配置文件路径时, 先从当前目录和用户目录/.config查找 <AppName>.yaml,
// 找不到时 或者加载错误时 使用内嵌的配置文件
func LoadConfig(cfgPath string, appName string) (*Config, error) {
	var data []byte
	var err error

	if cfgPath == "" {
		// 当用户没有指定配置文件路径时, 先从当前目录和用户目录/.config查找 <AppName>.yaml,
		// 找不到时 或者加载错误时 使用内嵌的配置文件
		defaultConfig := appName + ".yaml"
		cfgPath = findConfigPath(defaultConfig)
		if cfgPath != "" {
			data, err = os.ReadFile(cfgPath)
			if err != nil {
				fmt.Errorf("read found config %s error: %v", cfgPath, err)
				data = []byte(GetDefaultConfig())
			}
		} else {
			data = []byte(GetDefaultConfig())
		}
	} else {
		// 如果已指定配置文件 就从指定的配置中读取
		data, err = os.ReadFile(cfgPath)
		if err != nil {
			return nil, err
		}
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func findConfigPath(configName string) string {
	configPath := ""
	configPaths := []string{
		configName,
		filepath.Join(os.Getenv("HOME"), ".config", configName),
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			configPath = path
			break
		}
	}
	return configPath
}

// GetPreset 获取预设配置
func (c *Config) GetPreset(name string) (*PresetConfig, bool) {
	preset, exists := c.Presets[name]
	return &preset, exists
}

// GetDefaultConfig 获取内置默认配置文件内容（从嵌入文件获取）
func GetDefaultConfig() string {
	return embeds.GetConfig()
}

// GenDefaultConfig 生成默认配置文件到指定路径
func GenDefaultConfig(configPath string) error {
	defaultConfig := GetDefaultConfig()
	return os.WriteFile(configPath, []byte(defaultConfig), 0644)
}

// PrintPresetSummary 打印预设名称和描述
func PrintPresetSummary(cfg *Config) {
	fmt.Println("preset list:")
	fmt.Println("--------------------------------")
	// 遍历 map
	for name, preset := range cfg.Presets {
		// name 就是 map 的键（预设名称）
		// preset 就是 PresetConfig 结构体值
		fmt.Printf("- name: %s\n", name)
		fmt.Printf("  desc: %s\n", preset.Description)
		fmt.Println("--------------------------------")
	}
}
