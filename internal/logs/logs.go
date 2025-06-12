package logs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func LogFileSetup() {
	// Get current directory and move one level up
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("Unable to get working directory:", err)
	}
	parentDir := filepath.Dir(wd)
	logFilePath := filepath.Join(parentDir, "internal", "logs", "logs.txt")

	// Create/open log file
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Unable to create/open log file:", err)
	}
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func LogAndPrint(format string, args ...any) {
	log.Printf(format, args...)
	fmt.Printf(format, args...)
}
