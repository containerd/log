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

// Package log provides types and functions related to logging, passing
// loggers through a context, and attaching context to the logger.
//
// This package uses [log/slog] as the logging backend. It provides an
// [Entry] type for compatibility with code previously using logrus-style
// logging, while delegating all output to slog.
//
// Log level, format, and handler configuration should be done directly
// through [log/slog] using [slog.SetDefault] before using this package.
package log

import (
	"context"
	"log/slog"
)

// G is a shorthand for [GetLogger].
//
// We may want to define this locally to a package to get package tagged log
// messages.
var G = GetLogger

// L is the default logger.
var L = &Entry{}

type loggerKey struct{}

// WithLogger returns a new context with the provided logger. Use in
// combination with logger.WithField(s) for great effect.
func WithLogger(ctx context.Context, logger *Entry) context.Context {
	e := &Entry{
		logger: logger.logger,
		attr:   logger.attr,
		ctx:    ctx,
	}
	return context.WithValue(ctx, loggerKey{}, e)
}

// GetLogger retrieves the current logger from the context. If no logger is
// available, the default logger is returned.
func GetLogger(ctx context.Context) *Entry {
	if logger := ctx.Value(loggerKey{}); logger != nil {
		return logger.(*Entry)
	}
	return &Entry{
		logger: slog.Default(),
		ctx:    ctx,
	}
}
