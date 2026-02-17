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

// Package request contains utilities for working with RunFunctionRequests.
package request

import (
	"slices"

	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/crossplane/function-sdk-go/errors"
	v1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composed"
	"github.com/crossplane/function-sdk-go/resource/composite"
)

// GetInput from the supplied request. Input is loaded into the supplied object.
func GetInput(req *v1.RunFunctionRequest, into runtime.Object) error {
	return errors.Wrapf(resource.AsObject(req.GetInput(), into), "cannot get function input %T from %T", into, req)
}

// GetContextKey gets context from the supplied key.
func GetContextKey(req *v1.RunFunctionRequest, key string) (*structpb.Value, bool) {
	f := req.GetContext().GetFields()
	if f == nil {
		return nil, false
	}
	v, ok := f[key]
	return v, ok
}

// GetObservedCompositeResource from the supplied request.
func GetObservedCompositeResource(req *v1.RunFunctionRequest) (*resource.Composite, error) {
	xr := &resource.Composite{
		Resource:          composite.New(),
		ConnectionDetails: req.GetObserved().GetComposite().GetConnectionDetails(),
	}

	if xr.ConnectionDetails == nil {
		xr.ConnectionDetails = make(resource.ConnectionDetails)
	}

	err := resource.AsObject(req.GetObserved().GetComposite().GetResource(), xr.Resource)
	return xr, err
}

// GetObservedComposedResources from the supplied request.
func GetObservedComposedResources(req *v1.RunFunctionRequest) (map[resource.Name]resource.ObservedComposed, error) {
	ocds := map[resource.Name]resource.ObservedComposed{}
	for name, r := range req.GetObserved().GetResources() {
		ocd := resource.ObservedComposed{Resource: composed.New(), ConnectionDetails: r.GetConnectionDetails()}

		if ocd.ConnectionDetails == nil {
			ocd.ConnectionDetails = make(resource.ConnectionDetails)
		}

		if err := resource.AsObject(r.GetResource(), ocd.Resource); err != nil {
			return nil, err
		}
		ocds[resource.Name(name)] = ocd
	}
	return ocds, nil
}

// GetDesiredCompositeResource from the supplied request.
func GetDesiredCompositeResource(req *v1.RunFunctionRequest) (*resource.Composite, error) {
	xr := &resource.Composite{
		Resource:          composite.New(),
		ConnectionDetails: req.GetDesired().GetComposite().GetConnectionDetails(),
	}

	if xr.ConnectionDetails == nil {
		xr.ConnectionDetails = make(resource.ConnectionDetails)
	}

	err := resource.AsObject(req.GetDesired().GetComposite().GetResource(), xr.Resource)
	return xr, err
}

// GetDesiredComposedResources from the supplied request.
func GetDesiredComposedResources(req *v1.RunFunctionRequest) (map[resource.Name]*resource.DesiredComposed, error) {
	dcds := map[resource.Name]*resource.DesiredComposed{}
	for name, r := range req.GetDesired().GetResources() {
		dcd := &resource.DesiredComposed{Resource: composed.New()}
		if err := resource.AsObject(r.GetResource(), dcd.Resource); err != nil {
			return nil, err
		}
		switch r.GetReady() {
		case v1.Ready_READY_UNSPECIFIED:
			dcd.Ready = resource.ReadyUnspecified
		case v1.Ready_READY_TRUE:
			dcd.Ready = resource.ReadyTrue
		case v1.Ready_READY_FALSE:
			dcd.Ready = resource.ReadyFalse
		}
		dcds[resource.Name(name)] = dcd
	}
	return dcds, nil
}

// GetRequiredResources from the supplied request using the new required_resources field.
func GetRequiredResources(req *v1.RunFunctionRequest) (map[string][]resource.Required, error) {
	out := make(map[string][]resource.Required, len(req.GetRequiredResources()))
	for name, ers := range req.GetRequiredResources() {
		out[name] = []resource.Required{}
		for _, i := range ers.GetItems() {
			r := &resource.Required{Resource: &unstructured.Unstructured{}}
			if err := resource.AsObject(i.GetResource(), r.Resource); err != nil {
				return nil, err
			}
			out[name] = append(out[name], *r)
		}
	}
	return out, nil
}

// GetRequiredResource from the supplied request by name. The bool return value
// indicates whether Crossplane has resolved the requirement.
func GetRequiredResource(req *v1.RunFunctionRequest, name string) ([]resource.Required, bool, error) {
	if req.GetRequiredResources() == nil {
		return nil, false, nil
	}
	rrs, ok := req.GetRequiredResources()[name]
	if !ok {
		return nil, false, nil
	}
	out := make([]resource.Required, 0, len(rrs.GetItems()))
	for _, i := range rrs.GetItems() {
		r := &resource.Required{Resource: &unstructured.Unstructured{}}
		if err := resource.AsObject(i.GetResource(), r.Resource); err != nil {
			return nil, true, err
		}
		out = append(out, *r)
	}
	return out, true, nil
}

// GetExtraResources from the supplied request using the deprecated extra_resources field.
//
// Deprecated: Use GetRequiredResources.
func GetExtraResources(req *v1.RunFunctionRequest) (map[string][]resource.Required, error) {
	out := make(map[string][]resource.Required, len(req.GetExtraResources()))
	for name, ers := range req.GetExtraResources() {
		out[name] = []resource.Required{}
		for _, i := range ers.GetItems() {
			r := &resource.Required{Resource: &unstructured.Unstructured{}}
			if err := resource.AsObject(i.GetResource(), r.Resource); err != nil {
				return nil, err
			}
			out[name] = append(out[name], *r)
		}
	}
	return out, nil
}

// AdvertisesCapabilities returns true if Crossplane advertises its capabilities
// in the request metadata. Crossplane v2.2 and later advertise capabilities. If
// this returns false, the calling Crossplane predates capability advertisement
// and HasCapability will always return False, even for features the older
// Crossplane does support.
func AdvertisesCapabilities(req *v1.RunFunctionRequest) bool {
	return HasCapability(req, v1.Capability_CAPABILITY_CAPABILITIES)
}

// HasCapability returns true if Crossplane advertises the supplied capability
// in the request metadata. Functions can use this to determine whether
// Crossplane will honor certain fields in their response, or populate certain
// fields in their request.
//
// Use AdvertisesCapabilities to check whether Crossplane advertises its
// capabilities at all. If it doesn't, HasCapability always returns false even
// for features the older Crossplane does support.
func HasCapability(req *v1.RunFunctionRequest, c v1.Capability) bool {
	return slices.Contains(req.GetMeta().GetCapabilities(), c)
}

// GetRequiredSchemas from the supplied request. Returns all resolved schemas as
// a map of name to OpenAPI v3 schema protobuf Struct.
func GetRequiredSchemas(req *v1.RunFunctionRequest) map[string]*structpb.Struct {
	out := make(map[string]*structpb.Struct, len(req.GetRequiredSchemas()))
	for name, s := range req.GetRequiredSchemas() {
		out[name] = s.GetOpenapiV3()
	}
	return out
}

// GetRequiredSchema from the supplied request. Returns the OpenAPI v3 schema as
// a protobuf Struct. The bool return value indicates whether Crossplane has
// resolved the requirement. When ok is false, the schema was not yet resolved.
// When ok is true but the returned struct is nil, the schema was resolved but
// not found.
func GetRequiredSchema(req *v1.RunFunctionRequest, name string) (schema *structpb.Struct, ok bool) {
	schemas := req.GetRequiredSchemas()
	if schemas == nil {
		return nil, false
	}
	s, ok := schemas[name]
	if !ok {
		return nil, false
	}
	return s.GetOpenapiV3(), true
}

// GetCredentials from the supplied request.
func GetCredentials(req *v1.RunFunctionRequest, name string) (resource.Credentials, error) {
	cred, exists := req.GetCredentials()[name]
	if !exists {
		return resource.Credentials{}, errors.Errorf("%s: credential not found", name)
	}

	switch t := cred.GetSource().(type) {
	case *v1.Credentials_CredentialData:
		return resource.Credentials{Type: resource.CredentialsTypeData, Data: cred.GetCredentialData().GetData()}, nil
	default:
		return resource.Credentials{}, errors.Errorf("%s: not a supported credential source", t)
	}
}
