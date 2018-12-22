package app

import "github.com/XMatrixStudio/Violet.SDK.Go"

// Config 配置文件
type Config struct {
	Violet   violetSdk.Config `yaml:"Violet"`   // Violet配置
	Server   ServerConfig     `yaml:"Server"`   // 服务器配置
	Database DatabaseConfig   `yaml:"Database"` // 数据库配置
}

type ServerConfig struct {
	Port int `yaml:"Port"` // 端口
}

type DatabaseConfig struct {
	IP       string `yaml:"IP"`
	Port     int    `yaml:"Port"`
	Name     string `yaml:"Name"`
	Username string `yaml:"Username"`
	Password string `yaml:"Password"`
}
