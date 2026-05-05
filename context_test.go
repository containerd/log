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
	"context"
	"reflect"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestLoggerContext(t *testing.T) {
	const expected = "one"
	ctx := context.Background()
	ctx = WithLogger(ctx, G(ctx).WithField("test", expected))
	if actual := GetLogger(ctx).Data["test"]; actual != expected {
		t.Errorf("expected: %v, got: %v", expected, actual)
	}
	a := G(ctx)
	b := GetLogger(ctx)
	if !reflect.DeepEqual(a, b) || a != b {
		t.Errorf("should be the same: %+v, %+v", a, b)
	}
}

func TestSetLevel(t *testing.T) {
	for _, tc := range []struct {
		input    string
		expected Level
		wantErr  bool
	}{
		// Named levels
		{input: "trace", expected: TraceLevel},
		{input: "debug", expected: DebugLevel},
		{input: "info", expected: InfoLevel},
		{input: "warn", expected: WarnLevel},
		{input: "warning", expected: WarnLevel},
		{input: "error", expected: ErrorLevel},
		{input: "fatal", expected: FatalLevel},
		{input: "panic", expected: PanicLevel},
		{input: "INFO", expected: InfoLevel},
		{input: "DEBUG", expected: DebugLevel},

		// Exact numeric levels
		{input: "-8", expected: TraceLevel},
		{input: "-4", expected: DebugLevel},
		{input: "0", expected: InfoLevel},
		{input: "4", expected: WarnLevel},
		{input: "8", expected: ErrorLevel},
		{input: "10", expected: FatalLevel},
		{input: "12", expected: PanicLevel},

		// Nearest numeric levels (without going above)
		{input: "-10", expected: TraceLevel},
		{input: "-6", expected: DebugLevel},
		{input: "-5", expected: DebugLevel},
		{input: "-2", expected: InfoLevel},
		{input: "1", expected: WarnLevel},
		{input: "2", expected: WarnLevel},
		{input: "3", expected: WarnLevel},
		{input: "6", expected: ErrorLevel},
		{input: "7", expected: ErrorLevel},
		{input: "9", expected: FatalLevel},
		{input: "11", expected: PanicLevel},
		{input: "13", expected: PanicLevel},
		{input: "100", expected: PanicLevel},

		// Invalid
		{input: "bogus", wantErr: true},
		{input: "", wantErr: true},
	} {
		t.Run(tc.input, func(t *testing.T) {
			err := SetLevel(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if actual := GetLevel(); actual != tc.expected {
				t.Errorf("expected level %v, got %v", tc.expected, actual)
			}
		})
	}
}

func TestCompat(t *testing.T) {
	expected := Fields{
		"hello1": "world1",
		"hello2": "world2",
		"hello3": "world3",
	}

	l := G(context.TODO())
	l = l.WithFields(logrus.Fields{"hello1": "world1"})
	l = l.WithFields(Fields{"hello2": "world2"})
	l = l.WithFields(map[string]any{"hello3": "world3"})
	if !reflect.DeepEqual(Fields(l.Data), expected) {
		t.Errorf("expected: (%[1]T) %+[1]v, got: (%[2]T) %+[2]v", expected, l.Data)
	}

	l2 := L
	l2 = l2.WithFields(logrus.Fields{"hello1": "world1"})
	l2 = l2.WithFields(Fields{"hello2": "world2"})
	l2 = l2.WithFields(map[string]any{"hello3": "world3"})
	if !reflect.DeepEqual(Fields(l2.Data), expected) {
		t.Errorf("expected: (%[1]T) %+[1]v, got: (%[2]T) %+[2]v", expected, l2.Data)
	}
}
