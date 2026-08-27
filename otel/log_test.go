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
	"io"
	"testing"

	"github.com/containerd/log/otel"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"go.opentelemetry.io/otel/trace"
)

const expectedTraceIDStr = "0102030405060708090a0b0c0d0e0f10"

var (
	testTraceID = trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	testSpanID  = trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8}
)

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
