/*
Copyright 2024 The Crossplane Authors.

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

package response

import (
	"fmt"

	v1 "github.com/crossplane/function-sdk-go/proto/v1"
)

// ResultOption allows further customization of the result.
type ResultOption struct {
	result *v1.Result
}

// Fatal adds a fatal result to the supplied RunFunctionResponse.
// An event will be created for the Composite Resource.
// A fatal result cannot target the claim.
func Fatal(rsp *v1.RunFunctionResponse, err error) {
	newResult(rsp, v1.Severity_SEVERITY_FATAL, err.Error())
}

// Warning adds a warning result to the supplied RunFunctionResponse.
// An event will be created for the Composite Resource.
func Warning(rsp *v1.RunFunctionResponse, err error) *ResultOption {
	return newResult(rsp, v1.Severity_SEVERITY_WARNING, err.Error())
}

// Normal adds a normal result to the supplied RunFunctionResponse.
// An event will be created for the Composite Resource.
func Normal(rsp *v1.RunFunctionResponse, message string) *ResultOption {
	return newResult(rsp, v1.Severity_SEVERITY_NORMAL, message)
}

// Normalf adds a normal result to the supplied RunFunctionResponse.
// An event will be created for the Composite Resource.
func Normalf(rsp *v1.RunFunctionResponse, format string, a ...any) *ResultOption {
	return Normal(rsp, fmt.Sprintf(format, a...))
}

func newResult(rsp *v1.RunFunctionResponse, s v1.Severity, message string) *ResultOption {
	if rsp.GetResults() == nil {
		rsp.Results = make([]*v1.Result, 0, 1)
	}

	r := &v1.Result{
		Severity: s,
		Message:  message,
		Target:   v1.Target_TARGET_COMPOSITE.Enum(),
	}
	rsp.Results = append(rsp.GetResults(), r)

	return &ResultOption{result: r}
}

// TargetComposite updates the result and its event to target the composite
// resource.
func (o *ResultOption) TargetComposite() *ResultOption {
	o.result.Target = v1.Target_TARGET_COMPOSITE.Enum()
	return o
}

// TargetCompositeAndClaim updates the result and its event to target both the
// composite resource and claim.
func (o *ResultOption) TargetCompositeAndClaim() *ResultOption {
	o.result.Target = v1.Target_TARGET_COMPOSITE_AND_CLAIM.Enum()
	return o
}

// WithReason sets the reason field on the result and its event.
func (o *ResultOption) WithReason(reason string) *ResultOption {
	o.result.Reason = &reason
	return o
}
