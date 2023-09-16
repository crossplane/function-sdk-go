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

	"github.com/crossplane/crossplane-runtime/pkg/errors"

	"github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/resource"
)

// DefaultTTL is the default TTL for which a response can be cached.
const DefaultTTL = 1 * time.Minute

// To bootstraps a response to the supplied request. It automatically copies the
// desired state from the request.
func To(req *v1beta1.RunFunctionRequest, ttl time.Duration) *v1beta1.RunFunctionResponse {
	return &v1beta1.RunFunctionResponse{
		Meta: &v1beta1.ResponseMeta{
			Tag: req.GetMeta().GetTag(),
			Ttl: durationpb.New(ttl),
		},
		Desired: req.Desired,
	}
}

// SetDesiredCompositeResource sets the desired composite resource in the
// supplied response. The caller must be sure to avoid overwriting the desired
// state that may have been accumulated by previous Functions in the pipeline,
// unless they intend to.
func SetDesiredCompositeResource(rsp *v1beta1.RunFunctionResponse, xr *resource.Composite) error {
	if rsp.Desired == nil {
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
func SetDesiredComposedResources(rsp *v1beta1.RunFunctionResponse, dcds map[resource.Name]resource.DesiredComposed) error {
	if rsp.Desired == nil {
		rsp.Desired = &v1beta1.State{}
	}
	if rsp.Desired.Resources == nil {
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
func Fatal(rsp *v1beta1.RunFunctionResponse, err error) {
	if rsp.Results == nil {
		rsp.Results = make([]*v1beta1.Result, 0, 1)
	}
	rsp.Results = append(rsp.Results, &v1beta1.Result{
		Severity: v1beta1.Severity_SEVERITY_FATAL,
		Message:  err.Error(),
	})
}

// Warning adds a warning result to the supplied RunFunctionResponse.
func Warning(rsp *v1beta1.RunFunctionResponse, err error) {
	if rsp.Results == nil {
		rsp.Results = make([]*v1beta1.Result, 0, 1)
	}
	rsp.Results = append(rsp.Results, &v1beta1.Result{
		Severity: v1beta1.Severity_SEVERITY_WARNING,
		Message:  err.Error(),
	})
}

// Normal adds a normal result to the supplied RunFunctionResponse.
func Normal(rsp *v1beta1.RunFunctionResponse, message string) {
	if rsp.Results == nil {
		rsp.Results = make([]*v1beta1.Result, 0, 1)
	}
	rsp.Results = append(rsp.Results, &v1beta1.Result{
		Severity: v1beta1.Severity_SEVERITY_NORMAL,
		Message:  message,
	})
}

// Normalf adds a normal result to the supplied RunFunctionResponse.
func Normalf(rsp *v1beta1.RunFunctionResponse, format string, a ...any) {
	Normal(rsp, fmt.Sprintf(format, a...))
}
