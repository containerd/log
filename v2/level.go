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

import "log/slog"

// Level is a logging level.
type Level = slog.Level

// Supported log levels. These correspond to slog levels, with additional
// levels defined for trace, fatal, and panic for compatibility.
const (
	// LevelTrace designates finer-grained informational events than
	// [slog.LevelDebug].
	LevelTrace Level = slog.LevelDebug - 4

	// LevelDebug level. Usually only enabled when debugging. Very verbose
	// logging.
	LevelDebug Level = slog.LevelDebug

	// LevelInfo level. General operational entries about what's going on
	// inside the application.
	LevelInfo Level = slog.LevelInfo

	// LevelWarn level. Non-critical entries that deserve eyes.
	LevelWarn Level = slog.LevelWarn

	// LevelError level. Logs errors that should definitely be noted.
	LevelError Level = slog.LevelError

	// LevelFatal level. Provided for compatibility; maps to
	// slog.LevelError + 2.
	LevelFatal Level = slog.LevelError + 2

	// LevelPanic level. Provided for compatibility; maps to
	// slog.LevelError + 4.
	LevelPanic Level = slog.LevelError + 4
)
