package logger

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestCreate(t *testing.T) {
	logger := Create(false)
	if logger == nil {
		t.Error("Create() returned nil")
		return
	}

	if logger.logger == nil {
		t.Error("Logger not properly initialized")
	}

	if logger.quiet != false {
		t.Error("Expected quiet to be false")
	}
}

func TestCreateQuiet(t *testing.T) {
	logger := Create(true)
	if logger == nil {
		t.Error("Create() returned nil")
		return
	}

	if logger.quiet != true {
		t.Error("Expected quiet to be true")
	}
}

func TestInfoLogging(t *testing.T) {
	// Capture stderr output
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	logger := Create(false)
	testMessage := "test info message"
	logger.Info(testMessage)

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, testMessage) {
		t.Errorf("Expected output to contain '%s', got '%s'", testMessage, output)
	}
}

func TestInfoLoggingQuiet(t *testing.T) {
	// Capture stderr output
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	logger := Create(true) // quiet mode
	testMessage := "test info message"
	logger.Info(testMessage)

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if strings.Contains(output, testMessage) {
		t.Errorf("Expected no output in quiet mode, got '%s'", output)
	}
}

func TestInfofLogging(t *testing.T) {
	// Capture stderr output
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	logger := Create(false)
	logger.Infof("test formatted message: %d", 42)

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expectedMessage := "test formatted message: 42"
	if !strings.Contains(output, expectedMessage) {
		t.Errorf("Expected output to contain '%s', got '%s'", expectedMessage, output)
	}
}

func TestDebugLogging(t *testing.T) {
	// Capture stderr output
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	logger := Create(false)
	testMessage := "test debug message"
	logger.Debug(testMessage)

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, testMessage) {
		t.Errorf("Expected output to contain '%s', got '%s'", testMessage, output)
	}
}

func TestDebugfLogging(t *testing.T) {
	// Capture stderr output
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	logger := Create(false)
	logger.Debugf("debug formatted message: %s", "test")

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expectedMessage := "debug formatted message: test"
	if !strings.Contains(output, expectedMessage) {
		t.Errorf("Expected output to contain '%s', got '%s'", expectedMessage, output)
	}
}
