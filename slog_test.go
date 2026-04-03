/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package log

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

// setupSlogTest sets up UseSlogHook with a captured slog buffer and a separate
// buffer for the original logrus output. It restores all global state on cleanup.
func setupSlogTest(t *testing.T) (slogBuf, logrusBuf *bytes.Buffer) {
	t.Helper()

	// Save global state to restore later.
	oldLogger := L.Logger
	oldDefault := slog.Default()
	oldSlogOut := slogOut

	// Create a fresh logrus logger so we don't mutate the real global.
	logger := logrus.New()
	logger.SetLevel(logrus.TraceLevel)
	logrusBuf = &bytes.Buffer{}
	logger.SetOutput(logrusBuf)

	L = &Entry{
		Logger: logger,
		Data:   make(Fields, 6),
	}

	// Activate the slog hook — this redirects logrus output to io.Discard
	// and sets slogOut to the logger's original output.
	UseSlog()

	// Now install a slog handler that writes to our test buffer.
	slogBuf = &bytes.Buffer{}
	handler := slog.NewTextHandler(slogBuf, &slog.HandlerOptions{
		Level: slog.LevelDebug - 4, // capture all levels including trace
	})
	slog.SetDefault(slog.New(handler))

	t.Cleanup(func() {
		L = &Entry{
			Logger: oldLogger,
			Data:   make(Fields, 6),
		}
		slog.SetDefault(oldDefault)
		slogOut = oldSlogOut
	})

	return slogBuf, logrusBuf
}

func TestUseSlogHook(t *testing.T) {
	slogBuf, logrusBuf := setupSlogTest(t)

	L.Info("hello from L")

	slogOutput := slogBuf.String()
	logrusOutput := logrusBuf.String()

	if !strings.Contains(slogOutput, "hello from L") {
		t.Errorf("expected slog output to contain message, got: %s", slogOutput)
	}
	if logrusOutput != "" {
		t.Errorf("expected no logrus output, got: %s", logrusOutput)
	}
}

func TestUseSlogHookWithFields(t *testing.T) {
	slogBuf, logrusBuf := setupSlogTest(t)

	L.WithFields(Fields{
		"component": "test",
		"count":     42,
	}).Warn("something happened")

	slogOutput := slogBuf.String()

	if !strings.Contains(slogOutput, "something happened") {
		t.Errorf("expected slog output to contain message, got: %s", slogOutput)
	}
	if !strings.Contains(slogOutput, "component=test") {
		t.Errorf("expected slog output to contain component field, got: %s", slogOutput)
	}
	if !strings.Contains(slogOutput, "count=42") {
		t.Errorf("expected slog output to contain count field, got: %s", slogOutput)
	}
	if logrusBuf.Len() != 0 {
		t.Errorf("expected no logrus output, got: %s", logrusBuf.String())
	}
}

func TestUseSlogHookWithContext(t *testing.T) {
	slogBuf, logrusBuf := setupSlogTest(t)

	ctx := context.Background()
	logger := G(ctx).WithField("request_id", "abc123")
	ctx = WithLogger(ctx, logger)

	G(ctx).Info("context logger message")

	slogOutput := slogBuf.String()

	if !strings.Contains(slogOutput, "context logger message") {
		t.Errorf("expected slog output to contain message, got: %s", slogOutput)
	}
	if !strings.Contains(slogOutput, "request_id=abc123") {
		t.Errorf("expected slog output to contain request_id field, got: %s", slogOutput)
	}
	if logrusBuf.Len() != 0 {
		t.Errorf("expected no logrus output, got: %s", logrusBuf.String())
	}
}

func TestUseSlogHookLevels(t *testing.T) {
	slogBuf, logrusBuf := setupSlogTest(t)

	tests := []struct {
		name    string
		logFunc func(string, ...any)
		message string
	}{
		{"trace", L.Tracef, "trace-msg"},
		{"debug", L.Debugf, "debug-msg"},
		{"info", L.Infof, "info-msg"},
		{"warn", L.Warnf, "warn-msg"},
		{"error", L.Errorf, "error-msg"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			slogBuf.Reset()
			logrusBuf.Reset()

			tc.logFunc(tc.message)

			if !strings.Contains(slogBuf.String(), tc.message) {
				t.Errorf("expected slog output to contain %q, got: %s", tc.message, slogBuf.String())
			}
			if logrusBuf.Len() != 0 {
				t.Errorf("expected no logrus output, got: %s", logrusBuf.String())
			}
		})
	}
}

func TestSetFormatWithSlog(t *testing.T) {
	// SetFormat reconfigures the slog default handler to write to slogOut.
	// After SetFormat, logging through L should still go to slog (via slogOut),
	// and nothing should go to the logrus output (which is io.Discard).
	_, _ = setupSlogTest(t)

	// Replace slogOut with our own buffer so we can capture what SetFormat configures.
	var slogBuf bytes.Buffer
	slogOut = &slogBuf

	t.Run("text format", func(t *testing.T) {
		slogBuf.Reset()

		if err := SetFormat(TextFormat); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		L.Info("text format message")

		if !strings.Contains(slogBuf.String(), "text format message") {
			t.Errorf("expected slog output to contain message, got: %s", slogBuf.String())
		}
	})

	t.Run("json format", func(t *testing.T) {
		slogBuf.Reset()

		if err := SetFormat(JSONFormat); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		L.Info("json format message")

		slogOutput := slogBuf.String()
		if !strings.Contains(slogOutput, "json format message") {
			t.Errorf("expected slog output to contain message, got: %s", slogOutput)
		}
		if !strings.Contains(slogOutput, "{") {
			t.Errorf("expected JSON format output, got: %s", slogOutput)
		}
	})
}

func TestSetLevelWithSlog(t *testing.T) {
	slogBuf, _ := setupSlogTest(t)

	// Set level to warn — debug/info messages should be suppressed by slog.
	if err := SetLevel("warn"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Also reconfigure slog handler to use slogLevel (as SetFormat does).
	slogOut = slogBuf
	if err := SetFormat(TextFormat); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	slogBuf.Reset()
	L.Info("should be hidden")
	if slogBuf.Len() != 0 {
		t.Errorf("expected info message to be suppressed at warn level, got: %s", slogBuf.String())
	}

	slogBuf.Reset()
	L.Warn("should be visible")
	if !strings.Contains(slogBuf.String(), "should be visible") {
		t.Errorf("expected warn message to appear, got: %s", slogBuf.String())
	}

	// Raise level back to debug — info should now appear.
	if err := SetLevel("debug"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	slogBuf.Reset()
	L.Info("now visible")
	if !strings.Contains(slogBuf.String(), "now visible") {
		t.Errorf("expected info message to appear at debug level, got: %s", slogBuf.String())
	}
}

func TestLogrusToSlogLevel(t *testing.T) {
	tests := []struct {
		logrusLevel logrus.Level
		slogLevel   slog.Level
	}{
		{logrus.PanicLevel, slog.LevelError + 4},
		{logrus.FatalLevel, slog.LevelError + 2},
		{logrus.ErrorLevel, slog.LevelError},
		{logrus.WarnLevel, slog.LevelWarn},
		{logrus.InfoLevel, slog.LevelInfo},
		{logrus.DebugLevel, slog.LevelDebug},
		{logrus.TraceLevel, slog.LevelDebug - 4},
	}

	for _, tc := range tests {
		t.Run(tc.logrusLevel.String(), func(t *testing.T) {
			got := logrusToSlogLevel(tc.logrusLevel)
			if got != tc.slogLevel {
				t.Errorf("logrusToSlogLevel(%v) = %v, want %v", tc.logrusLevel, got, tc.slogLevel)
			}
		})
	}
}
