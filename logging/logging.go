/*
Copyright 2023 The Crossplane Authors.

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

// Package logging provides function's recommended logging interface.
package logging

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"

	"github.com/crossplane/function-sdk-go/errors"
)

// A Logger logs messages. Messages may be supplemented by structured data.
type FnLogger interface {
	// Info logs a message with optional structured data. Structured data must
	// be supplied as an array that alternates between string keys and values of
	// an arbitrary type. Use Info for messages that Crossplane operators are
	// very likely to be concerned with when running Crossplane.
	Info(msg string, keysAndValues ...any)

	// Debug logs a message with optional structured data. Structured data must
	// be supplied as an array that alternates between string keys and values of
	// an arbitrary type. Use Debug for messages that Crossplane operators or
	// developers may be concerned with when debugging Crossplane.
	Debug(msg string, keysAndValues ...any)

	// WithValues returns a Logger that will include the supplied structured
	// data with any subsequent messages it logs. Structured data must
	// be supplied as an array that alternates between string keys and values of
	// an arbitrary type.
	WithValues(keysAndValues ...any) FnLogger

	// Error logs a message with optional structured data. Structured data must
	// be supplied as an array that alternates between string keys and values of
	// an arbitrary type.
	// Use Error when logging fatal results of a function.
	Error(err error, msg string, keysAndValues ...any)
}

// NewNopLogger returns a Logger that does nothing.
func NewNopLogger() FnLogger { return nopFnLogger{} }

type nopFnLogger struct{}

func (l nopFnLogger) Info(_ string, _ ...any)           {}
func (l nopFnLogger) Debug(_ string, _ ...any)          {}
func (l nopFnLogger) Error(_ error, _ string, _ ...any) {}
func (l nopFnLogger) WithValues(_ ...any) FnLogger      { return nopFnLogger{} }

// NewLogrLogger returns a Logger that is satisfied by the supplied logr.Logger,
// which may be satisfied in turn by various logging implementations (Zap, klog,
// etc). Debug messages are logged at V(1).
func NewLogrLogger(l logr.Logger) FnLogger {
	return logrLogger{log: l}
}

type logrLogger struct {
	log logr.Logger
}

func (l logrLogger) Info(msg string, keysAndValues ...any) {
	l.log.Info(msg, keysAndValues...) //nolint:logrlint // False positive - logrlint thinks there's an odd number of args.
}

func (l logrLogger) Debug(msg string, keysAndValues ...any) {
	l.log.V(1).Info(msg, keysAndValues...) //nolint:logrlint // False positive - logrlint thinks there's an odd number of args.
}

func (l logrLogger) Error(err error, msg string, keysAndValues ...any) {
	l.log.Error(err, msg, keysAndValues...) //nolint:logrlint // False positive - logrlint thinks there's an odd number of args.
}

func (l logrLogger) WithValues(keysAndValues ...any) FnLogger {
	return logrLogger{log: l.log.WithValues(keysAndValues...)} //nolint:logrlint // False positive - logrlint thinks there's an odd number of args.
}

// NewLogger returns a new logger.
func NewLogger(debug bool) (FnLogger, error) {
	o := []zap.Option{zap.AddCallerSkip(1)}
	if debug {
		zl, err := zap.NewDevelopment(o...)
		return NewLogrLogger(zapr.NewLogger(zl)), errors.Wrap(err, "cannot create development zap logger")
	}
	zl, err := zap.NewProduction(o...)
	return NewLogrLogger(zapr.NewLogger(zl)), errors.Wrap(err, "cannot create production zap logger")
}
