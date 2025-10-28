package config

// PresetConfig 预设配置结构
type PresetConfig struct {
	Description string   `yaml:"description"`
	Include     []string `yaml:"include"`
	Exclude     []string `yaml:"exclude"`
	Ignored     []string `yaml:"ignored"`
}

// NewPresetConfig  带参数的构造函数
func NewPresetConfig(description string, include, exclude, ignored []string) *PresetConfig {
	return &PresetConfig{
		Description: description,
		Include:     copyStrSlice(include),
		Exclude:     copyStrSlice(exclude),
		Ignored:     copyStrSlice(ignored),
	}
}

// copyStrSlice safely copies a string slice, handling nil case.
func copyStrSlice(src []string) []string {
	if src == nil {
		return []string{}
	}
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}
