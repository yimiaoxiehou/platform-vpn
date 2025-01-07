package log

import (
	_log "log"
	"os"
	"time"
)

type Logger struct {
	items []*LogItem
}

type LogItem struct {
	Level   string
	Message string
	Time    time.Time
}

var _logger *Logger

func init() {
	_logger = &Logger{items: make([]*LogItem, 0)}
}

// 创建并返回日志记录器
func NewLogger() *Logger {
	if _logger == nil {
		_logger = &Logger{items: make([]*LogItem, 0)}
	}
	return _logger
}

func (log *Logger) GetLogs() []*LogItem {
	return log.items
}

func (log *Logger) Print(message string) {
	log.Info(message)
}

func (log *Logger) Trace(message string) {
	log.addItem("TRACE", message)
}

func (log *Logger) Debug(message string) {
	log.addItem("DEBUG", message)
}

func (log *Logger) Info(message string) {
	log.addItem("INFO", message)
}

func (log *Logger) Warning(message string) {
	log.addItem("WARNING", message)
}

func (log *Logger) Error(message string) {
	log.addItem("ERROR", message)
}

func (log *Logger) Fatal(message string) {
	log.addItem("FATAL", message)
	file, _ := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()
	for _, item := range log.items {
		file.WriteString(item.Time.Format("2024-01-01 12:00:00") + " " + item.Level + " " + item.Message + "\n")
	}
	os.Exit(1)
}

func (log *Logger) addItem(level string, message string) {
	_log.Print(message)
	log.items = append(log.items, &LogItem{Level: level, Message: message, Time: time.Now()})
}
