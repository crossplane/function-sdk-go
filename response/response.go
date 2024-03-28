/*
Copyright 2021 The Crossplane Authors.

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

// Package response contains utilities for working with RunFunctionResponses.
package response

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/crossplane/function-sdk-go/errors"
	"github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/resource"
)

// DefaultTTL is the default TTL for which a response can be cached.
const DefaultTTL = 1 * time.Minute

// A Result of running a Function.
type Result struct {
	result *v1beta1.Result
}

// To bootstraps a response to the supplied request. It automatically copies the
// desired state from the request.
func To(req *v1beta1.RunFunctionRequest, ttl time.Duration) *v1beta1.RunFunctionResponse {
	return &v1beta1.RunFunctionResponse{
		Meta: &v1beta1.ResponseMeta{
			Tag: req.GetMeta().GetTag(),
			Ttl: durationpb.New(ttl),
		},
		Desired: req.GetDesired(),
		Context: req.GetContext(),
	}
}

// SetContextKey sets context to the supplied key.
func SetContextKey(rsp *v1beta1.RunFunctionResponse, key string, v *structpb.Value) {
	if rsp.GetContext().GetFields() == nil {
		rsp.Context = &structpb.Struct{Fields: make(map[string]*structpb.Value)}
	}
	rsp.Context.Fields[key] = v
}

// SetDesiredCompositeResource sets the desired composite resource in the
// supplied response. The caller must be sure to avoid overwriting the desired
// state that may have been accumulated by previous Functions in the pipeline,
// unless they intend to.
func SetDesiredCompositeResource(rsp *v1beta1.RunFunctionResponse, xr *resource.Composite) error {
	if rsp.GetDesired() == nil {
		rsp.Desired = &v1beta1.State{}
	}
	s, err := resource.AsStruct(xr.Resource)
	rsp.Desired.Composite = &v1beta1.Resource{Resource: s, ConnectionDetails: xr.ConnectionDetails}
	return errors.Wrapf(err, "cannot convert %T to desired composite resource", xr.Resource)
}

// SetDesiredComposedResources sets the desired composed resources in the
// supplied response. The caller must be sure to avoid overwriting the desired
// state that may have been accumulated by previous Functions in the pipeline,
// unless they intend to.
func SetDesiredComposedResources(rsp *v1beta1.RunFunctionResponse, dcds map[resource.Name]*resource.DesiredComposed) error {
	if rsp.GetDesired() == nil {
		rsp.Desired = &v1beta1.State{}
	}
	if rsp.GetDesired().GetResources() == nil {
		rsp.Desired.Resources = map[string]*v1beta1.Resource{}
	}
	for name, dcd := range dcds {
		s, err := resource.AsStruct(dcd.Resource)
		if err != nil {
			return err
		}
		r := &v1beta1.Resource{Resource: s}
		switch dcd.Ready {
		case resource.ReadyUnspecified:
			r.Ready = v1beta1.Ready_READY_UNSPECIFIED
		case resource.ReadyFalse:
			r.Ready = v1beta1.Ready_READY_FALSE
		case resource.ReadyTrue:
			r.Ready = v1beta1.Ready_READY_TRUE
		}
		rsp.Desired.Resources[string(name)] = r
	}
	return nil
}

// Fatal adds a fatal result to the supplied RunFunctionResponse.
// A corresponding event will be created for the Composite Resource.
func Fatal(rsp *v1beta1.RunFunctionResponse, err error) *Result {
	if rsp.GetResults() == nil {
		rsp.Results = make([]*v1beta1.Result, 0, 1)
	}
	result := &v1beta1.Result{
		Severity: v1beta1.Severity_SEVERITY_FATAL,
		Message:  err.Error(),
		Target:   v1beta1.Target_TARGET_COMPOSITE,
	}
	rsp.Results = append(rsp.GetResults(), result)

	return &Result{
		result: result,
	}
}

// Warning adds a warning result to the supplied RunFunctionResponse.
// A corresponding event will be created for the Composite Resource.
func Warning(rsp *v1beta1.RunFunctionResponse, err error) *Result {
	if rsp.GetResults() == nil {
		rsp.Results = make([]*v1beta1.Result, 0, 1)
	}
	result := &v1beta1.Result{
		Severity: v1beta1.Severity_SEVERITY_WARNING,
		Message:  err.Error(),
		Target:   v1beta1.Target_TARGET_COMPOSITE,
	}
	rsp.Results = append(rsp.GetResults(), result)

	return &Result{
		result: result,
	}
}

// Normal adds a normal result to the supplied RunFunctionResponse.
// A corresponding event will be created for the Composite Resource.
func Normal(rsp *v1beta1.RunFunctionResponse, message string) *Result {
	if rsp.GetResults() == nil {
		rsp.Results = make([]*v1beta1.Result, 0, 1)
	}
	result := &v1beta1.Result{
		Severity: v1beta1.Severity_SEVERITY_NORMAL,
		Message:  message,
		Target:   v1beta1.Target_TARGET_COMPOSITE,
	}
	rsp.Results = append(rsp.GetResults(), result)

	return &Result{
		result: result,
	}
}

// Normalf adds a normal result to the supplied RunFunctionResponse.
// A corresponding event will be created for the Composite Resource.
func Normalf(rsp *v1beta1.RunFunctionResponse, format string, a ...any) *Result {
	return Normal(rsp, fmt.Sprintf(format, a...))
}

// TargetComposite configures the Composite to receive any events or conditions
// generated by this result.
func (o *Result) TargetComposite() *Result {
	o.result.Target = v1beta1.Target_TARGET_COMPOSITE
	return o
}

// TargetCompositeAndClaim configures both the Claim and Composite to receive
// any events or conditions generated by this result.
func (o *Result) TargetCompositeAndClaim() *Result {
	o.result.Target = v1beta1.Target_TARGET_COMPOSITE_AND_CLAIM
	return o
}

// ConditionTrue configures the result to create a condition on the targeted
// objects with the status set to true.
func (o *Result) ConditionTrue(t string, r string) *Result {
	o.result.Condition = &v1beta1.Condition{
		Type:   t,
		Status: v1beta1.Status_STATUS_TRUE,
		Reason: r,
	}
	return o
}

// ConditionFalse configures the result to create a condition on the targeted
// objects with the status set to false.
func (o *Result) ConditionFalse(t string, r string) *Result {
	o.result.Condition = &v1beta1.Condition{
		Type:   t,
		Status: v1beta1.Status_STATUS_FALSE,
		Reason: r,
	}
	return o
}

// ConditionUnknown configures the result to create a condition on the targeted
// objects with the status set to unknown.
func (o *Result) ConditionUnknown(t string, r string) *Result {
	o.result.Condition = &v1beta1.Condition{
		Type:   t,
		Status: v1beta1.Status_STATUS_UNKNOWN,
		Reason: r,
	}
	return o
}
