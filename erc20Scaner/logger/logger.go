package logger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/33cn/externaldb/erc20Scaner/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

// InitLogger 根据日志配置初始化日志系统
// 如果初始化失败，会返回错误，调用方应该处理错误并退出程序
// appName 是程序名（如 "rpc" 或 "scanner"），用于在日志文件名前添加前缀，避免多个程序使用同一配置时日志文件冲突
func InitLogger(logCfg config.LogConfig, appName string) (*slog.Logger, error) {
	logLevel := parseLogLevel(logCfg.Level)

	// 确定日志文件路径
	logFile := logCfg.File
	if logFile == "" {
		logFile = "./logs/app.log"
	}

	// 如果提供了程序名，在文件名前添加前缀
	if appName != "" {
		logDir := filepath.Dir(logFile)
		logBase := filepath.Base(logFile)
		// 如果文件名已经包含程序名前缀，不再重复添加
		if !strings.HasPrefix(logBase, appName+"_") {
			logFile = filepath.Join(logDir, appName+"_"+logBase)
		}
	}

	// 确保日志目录存在
	logDir := filepath.Dir(logFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory %s: %w", logDir, err)
	}

	// 获取日志轮转配置，使用默认值如果未设置
	maxSize := logCfg.MaxSize
	if maxSize <= 0 {
		maxSize = 100 // 默认 100 MB
	}
	maxBackups := logCfg.MaxBackups
	if maxBackups <= 0 {
		maxBackups = 5 // 默认保留 5 个文件
	}
	maxAge := logCfg.MaxAge
	if maxAge <= 0 {
		maxAge = 30 // 默认保留 30 天
	}

	// 配置日志轮转
	rotateWriter := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    maxSize,    // 单位：MB
		MaxBackups: maxBackups, // 保留旧日志文件的最大数量
		MaxAge:     maxAge,     // 单位：天
		Compress:   true,       // 是否压缩轮转后的旧日志文件
		LocalTime:  true,       // 使用本地时间
	}

	// 创建 JSON handler
	handler := slog.NewJSONHandler(rotateWriter, &slog.HandlerOptions{
		Level: logLevel,
	})

	return slog.New(handler), nil
}

// parseLogLevel 解析日志级别字符串为 slog.Level
func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		// 默认使用 info 级别
		return slog.LevelInfo
	}
}

// MaskDSN 隐藏 DSN 中的密码
// 将 user:password@host 格式转换为 user:***@host
func MaskDSN(dsn string) string {
	parts := strings.Split(dsn, "@")
	if len(parts) > 0 {
		userPass := strings.Split(parts[0], ":")
		if len(userPass) == 2 {
			return userPass[0] + ":***@" + strings.Join(parts[1:], "@")
		}
	}
	return dsn
}
