// config/config.go
package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	ClashProxy     string   `yaml:"clash_proxy"`
	NewsAPIKey     string   `yaml:"newsapi_key"`
	DeepSeekAPIKey string   `yaml:"deepseek_api_key"`
	OutputDir      string   `yaml:"output_dir"`
	Keywords       []string `yaml:"keywords"`
}

func LoadConfig(path string) *Config {
	data, err := os.ReadFile(path)
	if err != nil {
		panic("无法读取 config.yaml: " + err.Error())
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		panic("解析 config.yaml 失败: " + err.Error())
	}
	return &cfg
}
