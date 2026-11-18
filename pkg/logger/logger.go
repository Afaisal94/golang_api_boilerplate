package logger

import (
	"log"
	"os"
)

// Logger adalah wrapper tipis di atas log standar agar mudah diganti
// (mis. ke slog/zap) tanpa mengubah seluruh codebase.
type Logger struct {
	info  *log.Logger
	err   *log.Logger
	debug *log.Logger
}

// New membuat Logger dengan output ke stdout/stderr.
func New() *Logger {
	flags := log.Ldate | log.Ltime | log.Lshortfile
	return &Logger{
		info:  log.New(os.Stdout, "[INFO] ", flags),
		err:   log.New(os.Stderr, "[ERROR] ", flags),
		debug: log.New(os.Stdout, "[DEBUG] ", flags),
	}
}

// Default adalah instance global yang siap pakai.
var Default = New()

// Info mencatat pesan informasi.
func (l *Logger) Info(v ...any) {
	l.info.Println(v...)
}

// Infof mencatat pesan informasi dengan format.
func (l *Logger) Infof(format string, v ...any) {
	l.info.Printf(format, v...)
}

// Error mencatat pesan error.
func (l *Logger) Error(v ...any) {
	l.err.Println(v...)
}

// Errorf mencatat pesan error dengan format.
func (l *Logger) Errorf(format string, v ...any) {
	l.err.Printf(format, v...)
}

// Debug mencatat pesan debug.
func (l *Logger) Debug(v ...any) {
	l.debug.Println(v...)
}

// Debugf mencatat pesan debug dengan format.
func (l *Logger) Debugf(format string, v ...any) {
	l.debug.Printf(format, v...)
}

// Fatal mencatat pesan lalu keluar dengan status 1.
func (l *Logger) Fatal(v ...any) {
	l.err.Fatal(v...)
}

// Fatalf mencatat pesan format lalu keluar dengan status 1.
func (l *Logger) Fatalf(format string, v ...any) {
	l.err.Fatalf(format, v...)
}
