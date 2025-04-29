package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// 日志级别常量
const (
	LevelTrace   = "TRACE"
	LevelDebug   = "DEBUG"
	LevelInfo    = "INFO"
	LevelWarning = "WARNING"
	LevelError   = "ERROR"
	LevelFatal   = "FATAL"
)

// LogItem 表示一条日志记录
type LogItem struct {
	Level   string    // 日志级别
	Message string    // 日志消息
	Time    time.Time // 记录时间
}

// Logger 日志记录器
type Logger struct {
	items    []*LogItem   // 日志记录列表
	w        io.Writer    // 输出写入器
	mu       sync.RWMutex // 互斥锁
	maxItems int          // 最大记录数
	logFile  *os.File     // 日志文件
	logPath  string       // 日志文件路径
	maxSize  int64        // 日志文件最大大小（字节）
}

var (
	_logger *Logger
	once    sync.Once
)

// NewLogger 创建并返回日志记录器的单例实例
func NewLogger() *Logger {
	once.Do(func() {
		_logger = &Logger{
			items:    make([]*LogItem, 0, 1000),
			w:        os.Stdout,
			maxItems: 10000,
			logPath:  "logs",
			maxSize:  10 * 1024 * 1024, // 10MB
		}

		// 确保日志目录存在
		if err := os.MkdirAll(_logger.logPath, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "创建日志目录失败: %v\n", err)
		}
	})
	return _logger
}

// Write 实现 io.Writer 接口
func (log *Logger) Write(p []byte) (n int, err error) {
	message := string(p)
	// 解析日志级别和消息
	level := LevelInfo
	if strings.Contains(message, "level=info") {
		level = LevelInfo
	} else if strings.Contains(message, "level=warning") {
		level = LevelWarning
	} else if strings.Contains(message, "level=error") {
		level = LevelError
	} else if strings.Contains(message, "level=debug") {
		level = LevelDebug
	} else if strings.Contains(message, "level=trace") {
		level = LevelTrace
	} else if strings.Contains(message, "level=fatal") {
		level = LevelFatal
	}
	timeEnd := strings.Index(message, "msg=") + 5
	msg := message[timeEnd : len(message)-2]

	log.addItem(level, msg)
	return len(p), nil
}

// GetLogs 返回所有日志记录
func GetLogs() []*LogItem {
	logger := NewLogger()
	logger.mu.RLock()
	defer logger.mu.RUnlock()
	return append([]*LogItem{}, logger.items...)
}

func (log *Logger) addItem(level string, message string) {
	log.mu.Lock()
	defer log.mu.Unlock()

	// 创建日志项
	item := &LogItem{
		Level:   level,
		Message: message,
		Time:    time.Now(),
	}

	// 写入标准输出
	if log.w != nil {
		fmt.Fprintf(log.w, "[%s] %s %s\n",
			item.Time.Format("2006-01-02 15:04:05"),
			item.Level,
			item.Message)
	}

	// 限制内存中的日志数量
	if len(log.items) >= log.maxItems {
		log.items = log.items[1:]
	}
	log.items = append(log.items, item)
}

// 导出日志级别函数
func Trace(message string)   { NewLogger().addItem(LevelTrace, message) }
func Debug(message string)   { NewLogger().addItem(LevelDebug, message) }
func Info(message string)    { NewLogger().addItem(LevelInfo, message) }
func Warning(message string) { NewLogger().addItem(LevelWarning, message) }
func Error(message string)   { NewLogger().addItem(LevelError, message) }

// Fatal 记录致命错误并退出程序
func Fatal(message string) {
	logger := NewLogger()
	logger.addItem(LevelFatal, message)

	// 确保日志目录存在
	if err := os.MkdirAll(logger.logPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "创建日志目录失败: %v\n", err)
	}

	// 打开日志文件
	logFile := filepath.Join(logger.logPath, "app.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开日志文件失败: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// 写入所有日志记录
	logger.mu.RLock()
	for _, item := range logger.items {
		fmt.Fprintf(file, "[%s] %s %s\n",
			item.Time.Format("2006-01-02 15:04:05"),
			item.Level,
			item.Message)
	}
	logger.mu.RUnlock()

	os.Exit(1)
}
