package config

// PresetConfig 预设配置结构
type PresetConfig struct {
	Description string   `yaml:"description"`
	Allowed     []string `yaml:"allowed"`
	Removed     []string `yaml:"removed"`
	Ignored     []string `yaml:"ignored"`
}

// NewPresetConfig  带参数的构造函数
func NewPresetConfig(description string, allowed, removed, ignored []string) *PresetConfig {
	return &PresetConfig{
		Description: description,
		Allowed:     copyStrSlice(allowed),
		Removed:     copyStrSlice(removed),
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
