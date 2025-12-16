package utils

import (
	"log"
	"os"
)

// Logger с уровнями info и error
type Logger struct {
	info  *log.Logger
	error *log.Logger
}

// Создаём глобальный экземпляр
var Log = NewLogger()

func NewLogger() *Logger {
	return &Logger{
		info:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		error: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *Logger) Info(msg string) {
	l.info.Println(msg)
}

func (l *Logger) Error(msg string) {
	l.error.Println(msg)
}
