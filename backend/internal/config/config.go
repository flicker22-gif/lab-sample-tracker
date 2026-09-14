// Package config 集中管理进程级配置：全部来自环境变量，启动时一次性加载并校验。
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// 运行环境
const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

// Config 应用配置
type Config struct {
	Env  string // APP_ENV: development / production
	Port int    // PORT: HTTP 监听端口

	Database    DatabaseConfig
	MapStorage  string        // MAP_STORAGE_DIR: wafer map 原始文件存储目录
	SeedDemo    bool          // SEED_DEMO: 是否写入演示种子数据（默认 false，生产不得开启）
	LogLevel    string        // LOG_LEVEL: debug/info/warn/error
	LogFormat   string        // LOG_FORMAT: text/json（默认生产 json，其余 text）
	HTTP        HTTPConfig
	ShutdownSec time.Duration // SHUTDOWN_TIMEOUT_SEC: 优雅退出等待秒数
}

// DatabaseConfig 数据库连接与连接池配置
type DatabaseConfig struct {
	Host               string        // DB_HOST
	Port               int           // DB_PORT
	User               string        // DB_USER
	Password           string        // DB_PASSWORD
	Name               string        // DB_NAME
	SSLMode            string        // DB_SSLMODE
	TimeZone           string        // DB_TIMEZONE
	MaxOpenConns       int           // DB_MAX_OPEN_CONNS
	MaxIdleConns       int           // DB_MAX_IDLE_CONNS
	ConnMaxLifetimeMin int           // DB_CONN_MAX_LIFETIME_MINUTES
	ConnectRetries     int           // DB_CONNECT_RETRIES: 启动时等待数据库就绪的重试次数
	ConnectRetryWait   time.Duration // DB_CONNECT_RETRY_WAIT_MS: 每次重试间隔（毫秒）
}

// DSN 拼接 PostgreSQL 连接串
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode, d.TimeZone)
}

// HTTPConfig HTTP 服务器超时配置
type HTTPConfig struct {
	ReadHeaderTimeout time.Duration // HTTP_READ_HEADER_TIMEOUT_SEC
	ReadTimeout       time.Duration // HTTP_READ_TIMEOUT_SEC
	WriteTimeout      time.Duration // HTTP_WRITE_TIMEOUT_SEC
	IdleTimeout       time.Duration // HTTP_IDLE_TIMEOUT_SEC
}

// Load 从环境变量加载配置并做基础校验
func Load() (*Config, error) {
	env := strings.ToLower(getenv("APP_ENV", EnvDevelopment))
	switch env {
	case EnvDevelopment, EnvProduction, "test":
	default:
		return nil, fmt.Errorf("APP_ENV 非法: %q（可选 %s/%s）", env, EnvDevelopment, EnvProduction)
	}

	cfg := &Config{
		Env:  env,
		Port: getenvInt("PORT", 8080),
		Database: DatabaseConfig{
			Host:               getenv("DB_HOST", "localhost"),
			Port:               getenvInt("DB_PORT", 5432),
			User:               getenv("DB_USER", "tracker"),
			Password:           getenv("DB_PASSWORD", "tracker"),
			Name:               getenv("DB_NAME", "tracker"),
			SSLMode:            getenv("DB_SSLMODE", "disable"),
			TimeZone:           getenv("DB_TIMEZONE", "Asia/Shanghai"),
			MaxOpenConns:       getenvInt("DB_MAX_OPEN_CONNS", 20),
			MaxIdleConns:       getenvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetimeMin: getenvInt("DB_CONN_MAX_LIFETIME_MINUTES", 60),
			ConnectRetries:     getenvInt("DB_CONNECT_RETRIES", 30),
			ConnectRetryWait:   time.Duration(getenvInt("DB_CONNECT_RETRY_WAIT_MS", 2000)) * time.Millisecond,
		},
		MapStorage:  getenv("MAP_STORAGE_DIR", "./data/maps"),
		SeedDemo:    getenvBool("SEED_DEMO", false),
		LogLevel:    strings.ToLower(getenv("LOG_LEVEL", "info")),
		LogFormat:   strings.ToLower(getenv("LOG_FORMAT", "")),
		ShutdownSec: time.Duration(getenvInt("SHUTDOWN_TIMEOUT_SEC", 10)) * time.Second,
		HTTP: HTTPConfig{
			ReadHeaderTimeout: time.Duration(getenvInt("HTTP_READ_HEADER_TIMEOUT_SEC", 10)) * time.Second,
			ReadTimeout:       time.Duration(getenvInt("HTTP_READ_TIMEOUT_SEC", 60)) * time.Second,
			WriteTimeout:      time.Duration(getenvInt("HTTP_WRITE_TIMEOUT_SEC", 120)) * time.Second,
			IdleTimeout:       time.Duration(getenvInt("HTTP_IDLE_TIMEOUT_SEC", 120)) * time.Second,
		},
	}

	if cfg.LogFormat == "" {
		if env == EnvProduction {
			cfg.LogFormat = "json"
		} else {
			cfg.LogFormat = "text"
		}
	}
	switch cfg.LogFormat {
	case "json", "text":
	default:
		return nil, fmt.Errorf("LOG_FORMAT 非法: %q（可选 text/json）", cfg.LogFormat)
	}
	switch cfg.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return nil, fmt.Errorf("LOG_LEVEL 非法: %q（可选 debug/info/warn/error）", cfg.LogLevel)
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, fmt.Errorf("PORT 非法: %d", cfg.Port)
	}
	if cfg.Database.Name == "" || cfg.Database.User == "" {
		return nil, fmt.Errorf("DB_NAME 与 DB_USER 不能为空")
	}
	if cfg.SeedDemo && env == EnvProduction {
		// 不阻断启动，但必须在日志中显式可见；由调用方记录警告。
	}
	return cfg, nil
}

// GinMode 按环境推导 Gin 运行模式
func (c *Config) GinMode() string {
	if c.Env == EnvProduction {
		return "release"
	}
	return "debug"
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func getenvInt(k string, d int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return d
}

func getenvBool(k string, d bool) bool {
	if v := os.Getenv(k); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return d
}
