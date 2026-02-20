package utils

import (
	"fmt"
)

// Global Log variables
// Các biến Log toàn cục
// Global Log variables
// Các biến Log toàn cục
var VerboseMode bool

// Colors for console output
// Các màu cho đầu ra console
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
)

// SetupLogging initializes the log buffer.
// SetupLogging khởi tạo bộ đệm log.
func SetupLogging() {
	// No-op
}

// LogInfo prints a message to stdout.
// LogInfo in một tin nhắn ra stdout.
func LogInfo(msg string) {
	fmt.Println(msg)
}

// LogVerbose prints a message if VerboseMode is true.
// LogVerbose in một tin nhắn nếu VerboseMode là đúng.
func LogVerbose(msg string) {
	if VerboseMode {
		fmt.Println(Purple + "[VERBOSE] " + Cyan + msg + Reset)
	}
}

// WriteLog writes the accumulated log buffer to a file.
// WriteLog ghi bộ đệm log đã tích lũy vào một tệp.
func WriteLog() {
	// Deprecated: No file logging
	// Đã lỗi thời: Không ghi log ra tệp
}
