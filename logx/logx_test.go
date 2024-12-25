package logx

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestRegister(t *testing.T) {
	tests := []struct {
		name       string
		moduleName string
		wantName   string
	}{
		{
			name:       "empty module name should use default",
			moduleName: "",
			wantName:   "default",
		},
		{
			name:       "custom module name",
			moduleName: "test-module",
			wantName:   "test-module",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := Register(tt.moduleName, nil)
			if logger.moduleName != tt.wantName {
				t.Errorf("Register() moduleName = %v, want %v", logger.moduleName, tt.wantName)
			}
		})
	}
}

func TestGetLogger(t *testing.T) {
	// Register a test logger
	testModule := "test-get"
	Register(testModule, nil)

	tests := []struct {
		name       string
		moduleName string
		wantOK     bool
	}{
		{
			name:       "get existing logger",
			moduleName: testModule,
			wantOK:     true,
		},
		{
			name:       "get non-existing logger",
			moduleName: "non-existing",
			wantOK:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, ok := GetLogger(tt.moduleName)
			if ok != tt.wantOK {
				t.Errorf("GetLogger() ok = %v, want %v", ok, tt.wantOK)
			}
			if tt.wantOK && logger == nil {
				t.Error("GetLogger() logger is nil, want non-nil")
			}
		})
	}
}

func TestLogxWithContextExtractor(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create options with a context extractor
	opts := NewOptions(
		WithLevel(slog.LevelDebug),
		WithContextExtractor("request_id", func(ctx context.Context) string {
			if id, ok := ctx.Value("request_id").(string); ok {
				return id
			}
			return ""
		}),
		WithOutput(&buf),
	)

	// Create a new logger with the options
	logger := Register("test-context", opts)

	// Create a context with request_id
	ctx := context.WithValue(context.Background(), "request_id", "test-123")

	// Log a message
	logger.InfoContext(ctx, "test message")

	// Parse the log output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// Check if request_id was extracted and logged
	if requestID, ok := logEntry["request_id"].(string); !ok || requestID != "test-123" {
		t.Errorf("Expected request_id = test-123, got %v", requestID)
	}
}

func TestLogxWithFileOutput(t *testing.T) {
	// Create a temporary directory for test log file
	tmpDir, err := os.MkdirTemp("", "logx_test")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Temp dir: %v", tmpDir)
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "test.log")
	//how to get the log file handler
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	// Create options with file output
	opts := NewOptions(
		WithOutput(f),
		WithLevel(slog.LevelDebug),
	)

	// Create a new logger
	logger := Register("test-file", opts)

	// Log some messages
	logger.InfoContext(context.Background(), "test message")

	// Verify file was created and contains content
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}

	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}

	if len(content) == 0 {
		t.Error("Log file is empty")
	}
}

func TestDefault(t *testing.T) {
	// Clear any existing default logger
	moduleLoggers.Delete(defaultModuleName)

	// Get default logger
	logger := Default()
	if logger == nil {
		t.Error("Default() returned nil")
	}

	if logger.moduleName != defaultModuleName {
		t.Errorf("Default() moduleName = %v, want %v", logger.moduleName, defaultModuleName)
	}

	// Get it again, should be same instance
	logger2 := Default()
	if logger != logger2 {
		t.Error("Default() returned different instance on second call")
	}
}
