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
	"fmt"
	"log/slog"
	"testing"
)

func TestLoggerContext(t *testing.T) {
	ctx := context.Background()
	ctx = WithLogger(ctx, G(ctx).WithField("test", "one"))

	e := GetLogger(ctx)
	a := G(ctx)
	if e != a {
		t.Errorf("should be the same entry: %+v, %+v", e, a)
	}
}

func TestWithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	ctx := context.Background()
	ctx = WithLogger(ctx, &Entry{logger: logger, ctx: ctx})

	l := G(ctx)
	l = l.WithFields(Fields{"hello1": "world1"})
	l = l.WithFields(map[string]any{"hello2": "world2"})
	l.Info("test message")

	output := buf.String()
	for _, expected := range []string{"hello1=world1", "hello2=world2", "test message"} {
		if !bytes.Contains([]byte(output), []byte(expected)) {
			t.Errorf("expected output to contain %q, got: %s", expected, output)
		}
	}
}

func TestLogf(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: LevelTrace}))

	ctx := context.Background()
	ctx = WithLogger(ctx, &Entry{logger: logger, ctx: ctx})

	l := G(ctx)

	l.Tracef("trace %s", "msg")
	l.Debugf("debug %s", "msg")
	l.Infof("info %s", "msg")
	l.Warnf("warn %s", "msg")
	l.Errorf("error %s", "msg")

	output := buf.String()
	for _, expected := range []string{"trace msg", "debug msg", "info msg", "warn msg", "error msg"} {
		if !bytes.Contains([]byte(output), []byte(expected)) {
			t.Errorf("expected output to contain %q, got: %s", expected, output)
		}
	}
}

func TestWithError(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	ctx := context.Background()
	ctx = WithLogger(ctx, &Entry{logger: logger, ctx: ctx})

	l := G(ctx).WithError(fmt.Errorf("test error"))
	l.Info("something failed")

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("error")) {
		t.Errorf("expected output to contain error field, got: %s", output)
	}
	if !bytes.Contains([]byte(output), []byte("test error")) {
		t.Errorf("expected output to contain error message, got: %s", output)
	}
}
