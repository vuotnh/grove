package common

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
)

// GuestError là một struct để chứa lỗi tùy chỉnh
type GuestError struct {
	OriginalMessage string
}

func (e *GuestError) Error() string {
	return fmt.Sprintf("An error occurred communicating with the guest: %s.", e.OriginalMessage)
}

// LogAndRaise logAndRaise ghi log lỗi và trả về lỗi có chứa stack trace
func LogAndRaise(logFmt, excFmt string, fmtContent ...interface{}) {
	// Định dạng thông báo log và lỗi
	logMsg := logFmt
	excMsg := excFmt

	if len(fmtContent) > 0 {
		logMsg = fmt.Sprintf(logFmt, fmtContent...)
		excMsg = fmt.Sprintf(excFmt, fmtContent...)
	}

	// Ghi log lỗi
	log.Printf("ERROR: %s", logMsg)

	// Trả về lỗi kèm theo stack trace
	fmt.Fprintf(os.Stderr, "FATAL ERROR: %s\nExc: %s\n", excMsg, debug.Stack())
	os.Exit(1)
}
