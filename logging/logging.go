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
//
// Mainly a proxy for github.com/crossplane/crossplane-runtime/pkg/logging at
// the moment, but could diverge in the future if we see it fit.
package logging

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/crossplane/crossplane-runtime/pkg/logging"

	"github.com/crossplane/function-sdk-go/errors"
)

// A Logger logs messages. Messages may be supplemented by structured data.
type Logger logging.Logger

// NewNopLogger returns a Logger that does nothing.
func NewNopLogger() Logger { return logging.NewNopLogger() }

// NewLogrLogger returns a Logger that is satisfied by the supplied logr.Logger,
// which may be satisfied in turn by various logging implementations (Zap, klog,
// etc). Debug messages are logged at V(1).
func NewLogrLogger(l logr.Logger) Logger {
	return logging.NewLogrLogger(l)
}

// NewLogger returns a new logger.
func NewLogger(debug, timeEncodeISO8601 bool, addCallerSkip ...int) (Logger, error) {
	// default value for caller skip
	callerSkip := 1
	if len(addCallerSkip) > 0 {
		callerSkip = addCallerSkip[0]
	}

	o := []zap.Option{zap.AddCallerSkip(callerSkip)}
	if debug {
		zl, err := zap.NewDevelopment(o...)
		return NewLogrLogger(zapr.NewLogger(zl)), errors.Wrap(err, "cannot create development zap logger")
	}

	// If timeEncodeISO8601 is true, use ISO8601TimeEncoder for production logger.
	if timeEncodeISO8601 {
		pCfg := zap.NewProductionEncoderConfig()
		pCfg.EncodeTime = zapcore.ISO8601TimeEncoder

		p := zap.NewProductionConfig()
		p.EncoderConfig = pCfg
		zl, err := p.Build(o...)
		return NewLogrLogger(zapr.NewLogger(zl)), errors.Wrap(err, "cannot create production zap logger")
	}

	// If timeEncodeISO8601 is false, use the default EpochTimeEncoder for production logger.
	zl, err := zap.NewProduction(o...)
	return NewLogrLogger(zapr.NewLogger(zl)), errors.Wrap(err, "cannot create production zap logger")
}
