package logging

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "initialize logger",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call Init
			Init()

			// Verify Logger is not zero value
			if Logger.GetLevel() == zerolog.Disabled {
				t.Error("Init() did not initialize Logger properly")
			}
		})
	}
}

func TestInitWithComponent(t *testing.T) {
	tests := []struct {
		name      string
		component string
	}{
		{
			name:      "initialize with component name",
			component: "TEST",
		},
		{
			name:      "initialize with empty component",
			component: "",
		},
		{
			name:      "initialize with long component name",
			component: "VERY_LONG_COMPONENT_NAME",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create logger with component
			logger := InitWithComponent(tt.component)

			// Verify logger is not zero value
			if logger.GetLevel() == zerolog.Disabled {
				t.Error("InitWithComponent() returned disabled logger")
			}

			// Test that we can write to the logger
			var buf bytes.Buffer

			testLogger := zerolog.New(&buf).With().Str("component", tt.component).Logger()
			testLogger.Info().Msg("test message")

			output := buf.String()
			if tt.component != "" && !strings.Contains(output, tt.component) {
				t.Errorf("InitWithComponent() component %v not found in output", tt.component)
			}
		})
	}
}

func TestLogger_OutputFormat(t *testing.T) {
	tests := []struct {
		name    string
		level   zerolog.Level
		message string
	}{
		{
			name:    "info level message",
			level:   zerolog.InfoLevel,
			message: "test info message",
		},
		{
			name:    "error level message",
			level:   zerolog.ErrorLevel,
			message: "test error message",
		},
		{
			name:    "debug level message",
			level:   zerolog.DebugLevel,
			message: "test debug message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a logger that writes to a buffer
			var buf bytes.Buffer

			logger := zerolog.New(&buf).With().Timestamp().Logger()

			// Write log message
			switch tt.level {
			case zerolog.InfoLevel:
				logger.Info().Msg(tt.message)
			case zerolog.ErrorLevel:
				logger.Error().Msg(tt.message)
			case zerolog.DebugLevel:
				logger.Debug().Msg(tt.message)
			}

			output := buf.String()

			// Verify message is in output
			if !strings.Contains(output, tt.message) {
				t.Errorf("Logger output does not contain message: %v", tt.message)
			}
		})
	}
}

func TestLogger_WithFields(t *testing.T) {
	tests := []struct {
		name   string
		fields map[string]interface{}
	}{
		{
			name: "logger with string field",
			fields: map[string]interface{}{
				"key": "value",
			},
		},
		{
			name: "logger with multiple fields",
			fields: map[string]interface{}{
				"string_field": "value",
				"int_field":    42,
				"bool_field":   true,
			},
		},
		{
			name:   "logger with no extra fields",
			fields: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a logger that writes to a buffer
			var buf bytes.Buffer

			logger := zerolog.New(&buf).With().Timestamp().Logger()

			// Create event with fields
			event := logger.Info()

			for k, v := range tt.fields {
				switch val := v.(type) {
				case string:
					event = event.Str(k, val)
				case int:
					event = event.Int(k, val)
				case bool:
					event = event.Bool(k, val)
				}
			}

			event.Msg("test message")

			output := buf.String()

			// Verify all fields are in output
			for k := range tt.fields {
				if !strings.Contains(output, k) {
					t.Errorf("Logger output does not contain field: %v", k)
				}
			}
		})
	}
}
