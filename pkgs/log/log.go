package log

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type Logger struct {
	items []*LogItem
	w     io.Writer
}

type LogItem struct {
	Level   string
	Message string
	Time    time.Time
}

var _logger *Logger

func init() {
	_logger = &Logger{
		items: make([]*LogItem, 0),
		w:     os.Stdout,
	}
}

// Write implements io.Writer interface
func (log *Logger) Write(p []byte) (n int, err error) {
	message := string(p)
	msgStart := strings.Index(message, "msg=\"") + 5
	msgEnd := len(message) - 2
	msg := message[msgStart:msgEnd]
	if strings.Contains(msg, "level=trace") {
		Trace(msg)
	} else if strings.Contains(msg, "level=debug") {
		Debug(msg)
	} else if strings.Contains(msg, "level=info") {
		Info(msg)
	} else if strings.Contains(msg, "level=warning") {
		Warning(msg)
	} else if strings.Contains(msg, "level=error") {
		Error(msg)
	} else {
		Info(msg)
	}
	return len(msg), nil
}

// 创建并返回日志记录器
func NewLogger() *Logger {
	if _logger == nil {
		_logger = &Logger{
			items: make([]*LogItem, 0),
			w:     os.Stdout,
		}
	}
	return _logger
}

func GetLogs() []*LogItem {
	return NewLogger().items
}

func Trace(message string) {
	NewLogger().addItem("TRACE", message)
}

func Debug(message string) {
	NewLogger().addItem("DEBUG", message)
}

func Info(message string) {
	NewLogger().addItem("INFO", message)
}

func Warning(message string) {
	NewLogger().addItem("WARNING", message)
}

func Error(message string) {
	NewLogger().addItem("ERROR", message)
}

func Fatal(message string) {
	NewLogger().addItem("FATAL", message)
	file, _ := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()
	for _, item := range NewLogger().items {
		file.WriteString(item.Time.Format("2024-01-01 12:00:00") + " " + item.Level + " " + item.Message + "\n")
	}
	os.Exit(1)
}

func (log *Logger) addItem(level string, message string) {
	if log.w != nil {
		fmt.Fprintln(log.w, message)
	}
	log.items = append(log.items, &LogItem{Level: level, Message: message, Time: time.Now()})
}
