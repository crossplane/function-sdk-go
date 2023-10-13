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

// Package context contains utilities for working with Function context.
package context

// Well-known context keys.
const (
	// KeyEnvironment is the context key Crossplane sets to inject the
	// Composition Environment into Function context.
	//
	// https://github.com/crossplane/crossplane/blob/579702/design/one-pager-composition-environment.md
	//
	// THIS IS AN ALPHA FEATURE. Do not use it in production. It is not honored
	// unless the relevant Crossplane feature flag is enabled, and may be
	// changed or removed without notice.
	KeyEnvironment = "apiextensions.crossplane.io/environment"
)
