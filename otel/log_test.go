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

package otel_test

import (
	"context"
	"errors"
	"io"
	"slices"
	"testing"
	"time"

	"github.com/containerd/log"
	"github.com/containerd/log/otel"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const expectedTraceIDStr = "0102030405060708090a0b0c0d0e0f10"

var (
	testTraceID = trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	testSpanID  = trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8}
)

// testSpan is a minimal recording span used to test hook behavior without
// depending on the OpenTelemetry SDK.
type testSpan struct {
	trace.Span
	status codes.Code
}

func (s *testSpan) SpanContext() trace.SpanContext {
	return trace.NewSpanContext(trace.SpanContextConfig{TraceID: testTraceID, SpanID: testSpanID})
}

func (s *testSpan) IsRecording() bool                     { return true }
func (s *testSpan) AddEvent(string, ...trace.EventOption) {}
func (s *testSpan) SetStatus(code codes.Code, _ string)   { s.status = code }

func TestLogrusHookTraceID(t *testing.T) {
	tests := []struct {
		name        string
		enableOpt   bool
		nilContext  bool
		withSpan    bool
		expectedTID string
	}{
		{
			name:        "TraceIDInjected",
			enableOpt:   true,
			withSpan:    true,
			expectedTID: expectedTraceIDStr,
		},
		{
			name:      "TraceIDNotInjected_OptionDisabled",
			enableOpt: false,
			withSpan:  true,
		},
		{
			name:      "TraceIDNotInjected_NoSpan",
			enableOpt: true,
			withSpan:  false,
		},
		{
			name:       "TraceIDNotInjected_NoContext",
			enableOpt:  true,
			nilContext: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger := logrus.New()
			logger.SetOutput(io.Discard)
			logger.AddHook(otel.NewLogrusHook(otel.WithTraceIDField(tc.enableOpt)))
			testHook := test.NewLocal(logger)

			switch {
			case tc.withSpan:
				ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
					TraceID: testTraceID,
					SpanID:  testSpanID,
				}))
				logger.WithContext(ctx).Info("test")

			case tc.nilContext:
				logger.Info("test")

			default:
				logger.WithContext(context.Background()).Info("test")
			}

			entry := testHook.LastEntry()
			if entry == nil {
				t.Fatal("expected log entry")
			}

			traceID, ok := entry.Data["trace_id"]
			if tc.expectedTID != "" {
				if !ok {
					t.Fatal(`expected "trace_id" field`)
				}
				if traceID != tc.expectedTID {
					t.Errorf(`"trace_id" = %v; want %q`, traceID, tc.expectedTID)
				}
			} else if ok {
				t.Errorf(`unexpected "trace_id" field: %v`, traceID)
			}
		})
	}
}

// TestLogrusHookLevels verifies that [WithLevel] limits the levels handled by
// the hook while preserving all levels by default.
func TestLogrusHookLevels(t *testing.T) {
	tests := []struct {
		name string
		opts []otel.HookOpt
		want []log.Level
	}{
		{
			name: "default",
			want: []log.Level{
				log.PanicLevel,
				log.FatalLevel,
				log.ErrorLevel,
				log.WarnLevel,
				log.InfoLevel,
				log.DebugLevel,
				log.TraceLevel,
			},
		},
		{
			name: "warn",
			opts: []otel.HookOpt{
				otel.WithLevel(log.WarnLevel),
			},
			want: []log.Level{
				log.PanicLevel,
				log.FatalLevel,
				log.ErrorLevel,
				log.WarnLevel,
			},
		},
		{
			name: "error",
			opts: []otel.HookOpt{
				otel.WithLevel(log.ErrorLevel),
			},
			want: []log.Level{
				log.PanicLevel,
				log.FatalLevel,
				log.ErrorLevel,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hook := otel.NewLogrusHook(tc.opts...)
			if got := hook.Levels(); !slices.Equal(got, tc.want) {
				t.Errorf("Levels() = %v; want %v", got, tc.want)
			}
		})
	}
}

// TestLogrusHookErrorStatusLevel verifies that [WithErrorStatusLevel] marks
// spans as errors based on log severity and leaves span status unchanged by
// default.
func TestLogrusHookErrorStatusLevel(t *testing.T) {
	tests := []struct {
		name      string
		opts      []otel.HookOpt
		level     log.Level
		fields    log.Fields
		wantError bool
	}{
		{
			name:  "default",
			level: log.ErrorLevel,
		},
		{
			name: "below threshold",
			opts: []otel.HookOpt{
				otel.WithErrorStatusLevel(log.ErrorLevel),
			},
			level: log.WarnLevel,
		},
		{
			name: "at threshold",
			opts: []otel.HookOpt{
				otel.WithErrorStatusLevel(log.ErrorLevel),
			},
			level:     log.ErrorLevel,
			wantError: true,
		},
		{
			name: "above threshold",
			opts: []otel.HookOpt{
				otel.WithErrorStatusLevel(log.ErrorLevel),
			},
			level:     log.FatalLevel,
			wantError: true,
		},
		{
			name: "error field below threshold",
			opts: []otel.HookOpt{
				otel.WithErrorStatusLevel(log.ErrorLevel),
			},
			level: log.DebugLevel,
			fields: log.Fields{
				"error": errors.New("ignored"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			span := &testSpan{}
			ctx := trace.ContextWithSpan(context.Background(), span)

			hook := otel.NewLogrusHook(tc.opts...)
			err := hook.Fire(&log.Entry{
				Context: ctx,
				Data:    tc.fields,
				Level:   tc.level,
				Message: "message",
				Time:    time.Now(),
			})
			if err != nil {
				t.Fatal(err)
			}

			gotError := span.status == codes.Error
			if gotError != tc.wantError {
				t.Errorf("span error status = %v; want %v", gotError, tc.wantError)
			}
		})
	}
}
