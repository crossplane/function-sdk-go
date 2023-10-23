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
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/function-sdk-go/errors"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
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
func NewLogger(debug bool) (logging.Logger, error) {
	o := []zap.Option{zap.AddCallerSkip(1)}
	if debug {
		zl, err := zap.NewDevelopment(o...)
		return NewLogrLogger(zapr.NewLogger(zl)), errors.Wrap(err, "cannot create development zap logger")
	}
	zl, err := zap.NewProduction(o...)
	return NewLogrLogger(zapr.NewLogger(zl)), errors.Wrap(err, "cannot create production zap logger")
}
