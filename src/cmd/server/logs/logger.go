package logs

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/StepanKomis/Ticketa/src/cmd/server/env"
)

type Logger struct {
	logger *log.Logger
	file   *os.File
}
 
// NewLogger writes to stdout by default. Call AddWriter to add more destinations.
// prefix should be in lowercase and without []
func NewLogger(prefix string) (*Logger, error) {
		logPrefix := "[" + strings.ToUpper(prefix) + "] "
		
		filePath := fmt.Sprintf("/var/log/ticketa/%s.log", prefix)
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
        return nil, fmt.Errorf("failed to open log file: %w", err)
    }
    
		return &Logger{
        logger: log.New(io.MultiWriter(os.Stdout, f), logPrefix, log.LstdFlags|log.Lshortfile),
        file:   f,
    }, nil
}
 
// AddWriter adds an additional write destination (e.g. a file).
func (l *Logger) AddWriter(w io.Writer) {
	l.logger.SetOutput(io.MultiWriter(l.logger.Writer(), w))
}
 
func (l *Logger) Info(msg string)  { l.logger.Println("[INFO] " + msg) }
func (l *Logger) Error(msg string) { l.logger.Println("[ERROR] " + msg) }
func (l *Logger) Warn(msg string) { l.logger.Panicln("[WARN] " + msg) }

// Only shows the message when LOG_LEVEL env is set to debug
func (l *Logger) Debug(msg string)  {
	if strings.ToLower(env.Get("LOG_LEVEL", "info")) == "debug" {
  	l.logger.Println("[DEBUG] " + msg)
	}
}