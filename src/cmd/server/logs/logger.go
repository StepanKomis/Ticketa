package logs

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/StepanKomis/Ticketa/src/config"
)

type Logger struct {
	logger *log.Logger
	file   *os.File
	level  string
}

// NewLogger creates a logger that writes to stdout and a file under cfg.Logging.Dir.
// prefix should be lowercase and without [].
func NewLogger(prefix string, cfg *config.Config) (*Logger, error) {
	logPrefix := "[" + strings.ToUpper(prefix) + "] "

	filePath := fmt.Sprintf("%s/%s.log", cfg.Logging.Dir, prefix)
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Logger{
		logger: log.New(io.MultiWriter(os.Stdout, f), logPrefix, log.LstdFlags),
		file:   f,
		level:  strings.ToLower(cfg.Logging.Level),
	}, nil
}

// AddWriter adds an additional write destination (e.g. a file).
func (l *Logger) AddWriter(w io.Writer) {
	l.logger.SetOutput(io.MultiWriter(l.logger.Writer(), w))
}

func (l *Logger) Close() error { return l.file.Close() }

func (l *Logger) Info(msg string)  { l.logger.Println("[INFO] " + msg) }
func (l *Logger) Error(msg string) { l.logger.Println("[ERROR] " + msg) }
func (l *Logger) Warn(msg string)  { l.logger.Panicln("[WARN] " + msg) }
func (l *Logger) Fatal(msg string) {
	l.logger.Panicln("[FATAL] " + msg)
	os.Exit(1)
}

func (l *Logger) Infof(format string, args ...any)  { l.Info(fmt.Sprintf(format, args...)) }
func (l *Logger) Errorf(format string, args ...any) { l.Error(fmt.Sprintf(format, args...)) }
func (l *Logger) Warnf(format string, args ...any)  { l.Warn(fmt.Sprintf(format, args...)) }
func (l *Logger) Fatalf(format string, args ...any) { l.Fatal(fmt.Sprintf(format, args...)) }

func (l *Logger) Debug(msg string) {
	if l.level == "debug" {
		l.logger.Println("[DEBUG] " + msg)
	}
}

func (l *Logger) Debugf(format string, args ...any) { l.Debug(fmt.Sprintf(format, args...)) }
