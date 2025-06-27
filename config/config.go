package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// Config 配置结构体
type Config struct {
	Name       string // piko 客户端名称
	Remote     string // 远程 piko 服务器地址 (格式: host:port)
	ServerPort int    // piko 服务器端口
	Terminal   string // 终端类型
	AuthString string // 认证字符串
	SavePass   bool   // 是否保存密码
	Username   string // 用户名
	Password   string // 密码
	Timeout    int    // 超时时间（秒）
}

// NewConfig 创建新的配置实例
func NewConfig() *Config {
	return &Config{
		Name:       getEnvOrDefault("NAME", ""),
		Remote:     getEnvOrDefault("REMOTE", ""),
		ServerPort: 0,
		Terminal:   getEnvOrDefault("TERMINAL", ""), // 从环境变量读取终端类型
		SavePass:   getEnvBoolOrDefault("SAVE_PASS", false),
		Username:   getEnvOrDefault("USERNAME", ""),
		Password:   getEnvOrDefault("PASSWORD", ""),
		Timeout:    getEnvIntOrDefault("TIMEOUT", 30), // 默认30秒超时
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("客户端名称不能为空")
	}
	if c.Remote == "" {
		return fmt.Errorf("远程服务器地址不能为空")
	}
	return nil
}

// GetRemoteHost 获取远程主机地址
func (c *Config) GetRemoteHost() string {
	// 解析 remote 参数，格式: host:port
	parts := strings.Split(c.Remote, ":")
	if len(parts) >= 1 {
		return parts[0]
	}
	return "localhost"
}

// GetRemotePort 获取远程端口
func (c *Config) GetRemotePort() int {
	// 解析 remote 参数，格式: host:port
	parts := strings.Split(c.Remote, ":")
	if len(parts) >= 2 {
		if port, err := strconv.Atoi(parts[1]); err == nil {
			return port
		}
	}
	return 8088
}

// FindAvailablePort 查找可用端口，从8080开始
func (c *Config) FindAvailablePort() int {
	startPort := 8080
	for port := startPort; port < startPort+100; port++ {
		if isPortAvailable(port) {
			return port
		}
	}
	return startPort // 如果都不可用，返回默认端口
}

// isPortAvailable 检查端口是否可用
func isPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// getEnvOrDefault 获取环境变量或默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntOrDefault 获取整数环境变量或默认值
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBoolOrDefault 获取布尔环境变量或默认值
func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
