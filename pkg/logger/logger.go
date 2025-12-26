package logger

import (
	"log/slog"
	"os"
	"strings"
)

var L *slog.Logger

// Init 初始化全局日志器
func Init(level string) {
	var slogLevel slog.Level
	switch strings.ToUpper(level) {
	case "DEBUG":
		slogLevel = slog.LevelDebug
	case "INFO":
		slogLevel = slog.LevelInfo
	case "WARN":
		slogLevel = slog.LevelWarn
	case "ERROR":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	L = slog.New(handler)
	
	// 设置为默认 slog 日志器，这样第三方库也可以使用
	slog.SetDefault(L)
}

// Debug 调试日志
func Debug(msg string, args ...any) {
	L.Debug(msg, args...)
}

// Info 信息日志
func Info(msg string, args ...any) {
	L.Info(msg, args...)
}

// Warn 警告日志
func Warn(msg string, args ...any) {
	L.Warn(msg, args...)
}

// Error 错误日志
func Error(msg string, args ...any) {
	L.Error(msg, args...)
}
