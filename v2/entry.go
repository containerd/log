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
	"fmt"
	"log/slog"
)

// Entry is a logging entry. It contains all the fields passed with
// [Entry.WithFields]. It's finally logged when Trace, Debug, Info, Warn,
// or Error is called on it. These objects can be reused and passed around
// as much as you wish to avoid field duplication.
//
// Entry is for close compatibility with logrus and containerd/log,
// consider using slog directly for new packages.
//
// NOTE: while similar, this package is not compatible with logrus
// or containerd/log. It should be compatible for most uses but
// the interface is reduced and only supports slog capabilities.
// Some notable differences:
//   - No Fatal or Panic levels, no equivalent in slog
type Entry struct {
	logger *slog.Logger

	attr []slog.Attr

	ctx context.Context
}

// Fields type to pass to "WithFields".
type Fields = map[string]any

// WithError adds an error as single field to the Entry.
func (entry *Entry) WithError(err error) *Entry {
	return entry.WithField("error", err)
}

// WithContext adds a context to the Entry.
func (entry *Entry) WithContext(ctx context.Context) *Entry {
	return &Entry{
		logger: entry.logger,
		attr:   entry.attr,
		ctx:    ctx,
	}
}

// WithField adds a single field to the Entry.
func (entry *Entry) WithField(key string, value any) *Entry {
	return entry.WithFields(Fields{key: value})
}

// WithFields adds a map of fields to the Entry.
func (entry *Entry) WithFields(fields Fields) *Entry {
	attr := make([]slog.Attr, len(entry.attr), len(entry.attr)+len(fields))
	copy(attr, entry.attr)
	for k, v := range fields {
		attr = append(attr, slog.Any(k, v))
	}
	return &Entry{
		logger: entry.logger,
		attr:   attr,
		ctx:    entry.ctx,
	}
}

// Log logs a message at the given level.
func (entry *Entry) Log(level slog.Level, msg string) {
	logger := entry.logger
	if logger == nil {
		logger = slog.Default()
	}
	ctx := entry.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	logger.LogAttrs(ctx, level, msg, entry.attr...)
}

// Logf logs a formatted message at the given level.
func (entry *Entry) Logf(level slog.Level, format string, args ...any) {
	logger := entry.logger
	if logger == nil {
		logger = slog.Default()
	}
	if !logger.Enabled(entry.ctx, level) {
		return
	}
	ctx := entry.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	logger.LogAttrs(ctx, level, fmt.Sprintf(format, args...), entry.attr...)
}

// Trace logs a message at trace level (slog.LevelDebug-4).
func (entry *Entry) Trace(msg string) {
	entry.Log(LevelTrace, msg)
}

// Tracef logs a formatted message at trace level.
func (entry *Entry) Tracef(format string, args ...any) {
	entry.Logf(LevelTrace, format, args...)
}

// Debug logs a message at debug level.
func (entry *Entry) Debug(msg string) {
	entry.Log(slog.LevelDebug, msg)
}

// Debugf logs a formatted message at debug level.
func (entry *Entry) Debugf(format string, args ...any) {
	entry.Logf(slog.LevelDebug, format, args...)
}

// Info logs a message at info level.
func (entry *Entry) Info(msg string) {
	entry.Log(slog.LevelInfo, msg)
}

// Infof logs a formatted message at info level.
func (entry *Entry) Infof(format string, args ...any) {
	entry.Logf(slog.LevelInfo, format, args...)
}

// Warn logs a message at warn level.
func (entry *Entry) Warn(msg string) {
	entry.Log(slog.LevelWarn, msg)
}

// Warnf logs a formatted message at warn level.
func (entry *Entry) Warnf(format string, args ...any) {
	entry.Logf(slog.LevelWarn, format, args...)
}

// Error logs a message at error level.
func (entry *Entry) Error(msg string) {
	entry.Log(slog.LevelError, msg)
}

// Errorf logs a formatted message at error level.
func (entry *Entry) Errorf(format string, args ...any) {
	entry.Logf(slog.LevelError, format, args...)
}
