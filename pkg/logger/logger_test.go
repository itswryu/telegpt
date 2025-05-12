package logger

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/swryu/telegpt/pkg/config"
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

	// Initialize logger with debug level
	_ = Initialize(cfg)

	// Test different log levels
	Debug("Debug message")
	Info("Info message")
	Warn("Warn message")
	Error("Error message")

	// Restore stdout
	w.Close()
	os.Stdout = oldOutput

	// Read the output from the buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check that all messages were logged at debug level
	if !strings.Contains(output, "DEBUG: Debug message") {
		t.Errorf("Debug message not logged")
	}
	if !strings.Contains(output, "INFO: Info message") {
		t.Errorf("Info message not logged")
	}
	if !strings.Contains(output, "WARN: Warn message") {
		t.Errorf("Warn message not logged")
	}
	if !strings.Contains(output, "ERROR: Error message") {
		t.Errorf("Error message not logged")
	}

	// Now test a higher log level (info)
	buf.Reset()
	r, w, _ = os.Pipe()
	os.Stdout = w

	cfg.Logging.Level = "info"
	_ = Initialize(cfg)

	Debug("Debug message")
	Info("Info message")

	w.Close()
	os.Stdout = oldOutput

	io.Copy(&buf, r)
	output = buf.String()

	// Info level shouldn't include Debug messages
	if strings.Contains(output, "DEBUG: Debug message") {
		t.Errorf("Debug message was logged at info level")
	}
	if !strings.Contains(output, "INFO: Info message") {
		t.Errorf("Info message not logged at info level")
	}
}
