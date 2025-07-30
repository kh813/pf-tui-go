package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	logDir        = "~/.config/pf-tui"
	logFileName   = "pf-tui.log"
	maxBackups    = 30
	maxAgeDays    = 90
)

var (
	logger *log.Logger
)

func init() {
	setupLogging()
}

func setupLogging() {
	expandedLogDir := expandUser(logDir)
	if err := os.MkdirAll(expandedLogDir, 0755); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	logFilePath := filepath.Join(expandedLogDir, logFileName)

	// Configure lumberjack for log rotation
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    10, // megabytes
		MaxBackups: maxBackups,
		MaxAge:     maxAgeDays, // days
		Compress:   true,       // compress rotated files
	}

	// Set up the standard logger to write to lumberjack
	logger = log.New(lumberjackLogger, "", log.Ldate|log.Ltime|log.Lshortfile)

	// Perform log cleanup on startup
	go cleanupOldLogs(expandedLogDir)

	logger.Println("INFO: Logging initialized.")
}

func cleanupOldLogs(dir string) {
	cutoff := time.Now().AddDate(0, 0, -maxAgeDays)

	files, err := os.ReadDir(dir)
	if err != nil {
		logger.Printf("ERROR: Failed to read log directory for cleanup: %v", err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		info, err := file.Info()
		if err != nil {
			logger.Printf("WARN: Failed to get file info for %s: %v", file.Name(), err)
			continue
		}

		if info.ModTime().Before(cutoff) && (filepath.Ext(file.Name()) == ".gz" || filepath.Ext(file.Name()) == ".log") {
			filePath := filepath.Join(dir, file.Name())
			if err := os.Remove(filePath); err != nil {
				logger.Printf("ERROR: Failed to delete old log file %s: %v", filePath, err)
			} else {
				logger.Printf("INFO: Deleted old log file: %s", filePath)
			}
		}
	}
}

// expandUser expands the `~` symbol in a path to the user's home directory.
func expandUser(path string) string {
	if len(path) > 0 && path[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Fallback or error handling if home directory cannot be determined
			return path // Return original path if expansion fails
		}
		return filepath.Join(homeDir, path[1:])
	}
	return path
}

// Log functions for different levels
func LogInfo(format string, v ...interface{}) {
	logger.Printf("INFO: "+format, v...)
}

func LogWarn(format string, v ...interface{}) {
	logger.Printf("WARN: "+format, v...)
}

func LogError(format string, v ...interface{}) {
	logger.Printf("ERROR: "+format, v...)
}