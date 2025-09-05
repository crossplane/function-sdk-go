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
	"encoding/json"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/crossplane/function-sdk-go/errors"
	v1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/resource"
)

// DefaultTTL is the default TTL for which a response can be cached.
const DefaultTTL = 1 * time.Minute

// To bootstraps a response to the supplied request. It automatically copies the
// desired state from the request.
func To(req *v1.RunFunctionRequest, ttl time.Duration) *v1.RunFunctionResponse {
	return &v1.RunFunctionResponse{
		Meta: &v1.ResponseMeta{
			Tag: req.GetMeta().GetTag(),
			Ttl: durationpb.New(ttl),
		},
		Desired: req.GetDesired(),
		Context: req.GetContext(),
	}
}

// SetContextKey sets context to the supplied key.
func SetContextKey(rsp *v1.RunFunctionResponse, key string, v *structpb.Value) {
	if rsp.GetContext().GetFields() == nil {
		rsp.Context = &structpb.Struct{Fields: make(map[string]*structpb.Value)}
	}
	rsp.Context.Fields[key] = v
}

// SetDesiredCompositeResource sets the desired composite resource in the
// supplied response. The caller must be sure to avoid overwriting the desired
// state that may have been accumulated by previous Functions in the pipeline,
// unless they intend to.
func SetDesiredCompositeResource(rsp *v1.RunFunctionResponse, xr *resource.Composite) error {
	if rsp.GetDesired() == nil {
		rsp.Desired = &v1.State{}
	}
	s, err := resource.AsStruct(xr.Resource)
	r := &v1.Resource{Resource: s, ConnectionDetails: xr.ConnectionDetails}
	if err != nil {
		return errors.Wrapf(err, "cannot convert %T to desired composite resource", xr.Resource)
	}
	switch xr.Ready {
	case resource.ReadyUnspecified:
		r.Ready = v1.Ready_READY_UNSPECIFIED
	case resource.ReadyFalse:
		r.Ready = v1.Ready_READY_FALSE
	case resource.ReadyTrue:
		r.Ready = v1.Ready_READY_TRUE
	}
	rsp.Desired.Composite = r
	return nil
}

// SetDesiredComposedResources sets the desired composed resources in the
// supplied response. The caller must be sure to avoid overwriting the desired
// state that may have been accumulated by previous Functions in the pipeline,
// unless they intend to.
func SetDesiredComposedResources(rsp *v1.RunFunctionResponse, dcds map[resource.Name]*resource.DesiredComposed) error {
	if rsp.GetDesired() == nil {
		rsp.Desired = &v1.State{}
	}
	if rsp.GetDesired().GetResources() == nil {
		rsp.Desired.Resources = map[string]*v1.Resource{}
	}
	for name, dcd := range dcds {
		s, err := resource.AsStruct(dcd.Resource)
		if err != nil {
			return err
		}
		r := &v1.Resource{Resource: s}
		switch dcd.Ready {
		case resource.ReadyUnspecified:
			r.Ready = v1.Ready_READY_UNSPECIFIED
		case resource.ReadyFalse:
			r.Ready = v1.Ready_READY_FALSE
		case resource.ReadyTrue:
			r.Ready = v1.Ready_READY_TRUE
		}
		rsp.Desired.Resources[string(name)] = r
	}
	return nil
}

// SetDesiredResources sets the desired resources in the supplied response. The
// caller must be sure to avoid overwriting the desired state that may have been
// accumulated by previous Functions in the pipeline, unless they intend to.
func SetDesiredResources(rsp *v1.RunFunctionResponse, drs map[resource.Name]*unstructured.Unstructured) error {
	if rsp.GetDesired() == nil {
		rsp.Desired = &v1.State{}
	}
	if rsp.GetDesired().GetResources() == nil {
		rsp.Desired.Resources = map[string]*v1.Resource{}
	}
	for name, r := range drs {
		s, err := resource.AsStruct(r)
		if err != nil {
			return err
		}
		rsp.Desired.Resources[string(name)] = &v1.Resource{Resource: s}
	}
	return nil
}

// SetOutput sets the function's output. The supplied output must be marshalable
// as JSON. Only operation functions support setting output. If a composition
// function sets output it'll be ignored.
func SetOutput(rsp *v1.RunFunctionResponse, output any) error {
	j, err := json.Marshal(output)
	if err != nil {
		return errors.Wrap(err, "cannot marshal output to JSON")
	}

	rsp.Output = &structpb.Struct{}
	return errors.Wrap(protojson.Unmarshal(j, rsp.Output), "cannot unmarshal JSON to protobuf struct") //nolint:protogetter // It's a set.
}
