package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Logger struct {
	appLog     *log.Logger
	appFile    *os.File
	auditFile  *os.File
	dataDir    string
	logsDir    string
	sessionsDir string
}

func Init(dataDir string) (*Logger, error) {
	if dataDir == "" {
		return nil, fmt.Errorf("logging: data_dir is empty")
	}

	dirs := []string{
		filepath.Join(dataDir, "state"),
		filepath.Join(dataDir, "logs", "sessions"),
		filepath.Join(dataDir, "audit"),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return nil, fmt.Errorf("logging: create directory %s: %w", d, err)
		}
		// Verify writable by creating a temp file
		testFile := filepath.Join(d, ".write_test")
		if err := os.WriteFile(testFile, []byte{}, 0644); err != nil {
			return nil, fmt.Errorf("logging: directory %s is not writable: %w", d, err)
		}
		os.Remove(testFile)
	}

	logsDir := filepath.Join(dataDir, "logs")
	appLogPath := filepath.Join(logsDir, "app.log")
	appFile, err := os.OpenFile(appLogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("logging: open app.log: %w", err)
	}

	appLogger := log.New(appFile, "", log.LstdFlags|log.Lmsgprefix)

	auditPath := filepath.Join(dataDir, "audit", "audit.log")
	auditFile, err := os.OpenFile(auditPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("logging: open audit.log: %w", err)
	}

	return &Logger{
		appLog:      appLogger,
		appFile:     appFile,
		auditFile:   auditFile,
		dataDir:     dataDir,
		logsDir:     logsDir,
		sessionsDir: filepath.Join(dataDir, "logs", "sessions"),
	}, nil
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.appLog.Printf("[INFO] "+format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.appLog.Printf("[ERROR] "+format, v...)
}

func (l *Logger) Warn(format string, v ...interface{}) {
	l.appLog.Printf("[WARN] "+format, v...)
}

func (l *Logger) Debug(format string, v ...interface{}) {
	l.appLog.Printf("[DEBUG] "+format, v...)
}

func (l *Logger) Audit(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	fmt.Fprintf(l.auditFile, "%s\n", msg)
}

func (l *Logger) SessionLogWriter(sessionKey string) (io.WriteCloser, error) {
	safe := safeKey(sessionKey)
	path := filepath.Join(l.sessionsDir, safe+".log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("logging: open session log %s: %w", safe, err)
	}
	return f, nil
}

func (l *Logger) ReadSessionLog(sessionKey string, maxLines int) ([]string, error) {
	safe := safeKey(sessionKey)
	path := filepath.Join(l.sessionsDir, safe+".log")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	lines := splitLines(string(data))
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	return lines, nil
}

func (l *Logger) DataDir() string  { return l.dataDir }
func (l *Logger) StateDir() string { return filepath.Join(l.dataDir, "state") }

func (l *Logger) Close() error {
	l.appFile.Close()
	l.auditFile.Close()
	return nil
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	// Remove trailing empty line from split
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func safeKey(key string) string {
	r := strings.NewReplacer(
		":", "_",
		"\\", "_",
		"/", "_",
		" ", "_",
		"<", "_",
		">", "_",
		"|", "_",
		"?", "_",
		"*", "_",
	)
	return r.Replace(key)
}
