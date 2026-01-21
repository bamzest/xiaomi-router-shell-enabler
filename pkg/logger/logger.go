package logger

import (
	"fmt"
	"github.com/fatih/color"
	"log"
	"os"
)

// 日志级别
const (
	LevelDebug = iota
	LevelInfo
	LevelWarn
	LevelError
)

var (
	currentLevel = LevelInfo
	logger       = log.New(os.Stdout, "", 0)
	
	// 颜色输出
	debugColor = color.New(color.FgCyan)
	infoColor  = color.New(color.FgGreen)
	warnColor  = color.New(color.FgYellow)
	errorColor = color.New(color.FgRed, color.Bold)
)

// SetLevel 设置日志级别
func SetLevel(level int) {
	currentLevel = level
}

// Debug 输出调试日志
func Debug(format string, args ...interface{}) {
	if currentLevel <= LevelDebug {
		message := fmt.Sprintf(format, args...)
		logger.Println(debugColor.Sprintf("[DEBUG] %s", message))
	}
}

// Info 输出信息日志
func Info(format string, args ...interface{}) {
	if currentLevel <= LevelInfo {
		message := fmt.Sprintf(format, args...)
		logger.Println(infoColor.Sprintf("[INFO] %s", message))
	}
}

// Warn 输出警告日志
func Warn(format string, args ...interface{}) {
	if currentLevel <= LevelWarn {
		message := fmt.Sprintf(format, args...)
		logger.Println(warnColor.Sprintf("[WARN] %s", message))
	}
}

// Error 输出错误日志
func Error(format string, args ...interface{}) {
	if currentLevel <= LevelError {
		message := fmt.Sprintf(format, args...)
		logger.Println(errorColor.Sprintf("[ERROR] %s", message))
	}
}