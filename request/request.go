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
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/crossplane/function-sdk-go/errors"
	"github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composed"
	"github.com/crossplane/function-sdk-go/resource/composite"
)

// GetInput from the supplied request. Input is loaded into the supplied object.
func GetInput(req *v1beta1.RunFunctionRequest, into runtime.Object) error {
	return errors.Wrap(resource.AsObject(req.GetInput(), into), "cannot get Function input %T from %T, into, req")
}

// GetContextKey gets context from the supplied key.
func GetContextKey(req *v1beta1.RunFunctionRequest, key string) (*structpb.Value, bool) {
	f := req.GetContext().GetFields()
	if f == nil {
		return nil, false
	}
	v, ok := f[key]
	return v, ok
}

// GetObservedCompositeResource from the supplied request.
func GetObservedCompositeResource(req *v1beta1.RunFunctionRequest) (*resource.Composite, error) {
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
func GetObservedComposedResources(req *v1beta1.RunFunctionRequest) (map[resource.Name]resource.ObservedComposed, error) {
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
func GetDesiredCompositeResource(req *v1beta1.RunFunctionRequest) (*resource.Composite, error) {
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
func GetDesiredComposedResources(req *v1beta1.RunFunctionRequest) (map[resource.Name]*resource.DesiredComposed, error) {
	dcds := map[resource.Name]*resource.DesiredComposed{}
	for name, r := range req.GetDesired().GetResources() {
		dcd := &resource.DesiredComposed{Resource: composed.New()}
		if err := resource.AsObject(r.GetResource(), dcd.Resource); err != nil {
			return nil, err
		}
		switch r.GetReady() {
		case v1beta1.Ready_READY_UNSPECIFIED:
			dcd.Ready = resource.ReadyUnspecified
		case v1beta1.Ready_READY_TRUE:
			dcd.Ready = resource.ReadyTrue
		case v1beta1.Ready_READY_FALSE:
			dcd.Ready = resource.ReadyFalse
		}
		dcds[resource.Name(name)] = dcd
	}
	return dcds, nil
}

// GetExtraResources from the supplied request.
func GetExtraResources(req *v1beta1.RunFunctionRequest) (map[string][]resource.Extra, error) {
	out := make(map[string][]resource.Extra, len(req.GetExtraResources()))
	for name, ers := range req.GetExtraResources() {
		out[name] = []resource.Extra{}
		for _, i := range ers.GetItems() {
			r := &resource.Extra{Resource: &unstructured.Unstructured{}}
			if err := resource.AsObject(i.GetResource(), r.Resource); err != nil {
				return nil, err
			}
			out[name] = append(out[name], *r)
		}
	}
	return out, nil
}
