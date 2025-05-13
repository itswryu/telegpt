package logger

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/swryu/telegpt/pkg/config"
)

const (
	debugMessage = "Debug message"
	infoMessage  = "Info message"
	warnMessage  = "Warn message"
	errorMessage = "Error message"
)

func TestLogLevels(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	oldOutput := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create test config
	cfg := &config.Config{
		Logging: config.LoggingConfig{
			Level:   "debug",
			File:    "",
			Console: true,
		},
	}

	// Simplify test by using a single logger configuration
	cfg.Logging.Level = "debug"
	_ = Initialize(cfg)

	Debug(debugMessage)
	Info(infoMessage)
	Warn(warnMessage)
	Error(errorMessage)

	// Capture output
	w.Close()
	os.Stdout = oldOutput
	io.Copy(&buf, r)
	output := buf.String()

	// Check for log messages
	if !strings.Contains(output, debugMessage) {
		t.Errorf("Debug message not logged")
	}
	if !strings.Contains(output, infoMessage) {
		t.Errorf("Info message not logged")
	}
	if !strings.Contains(output, warnMessage) {
		t.Errorf("Warn message not logged")
	}
	if !strings.Contains(output, errorMessage) {
		t.Errorf("Error message not logged")
	}
}
