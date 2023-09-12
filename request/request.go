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
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/resource/unstructured/composed"
	"github.com/crossplane/crossplane-runtime/pkg/resource/unstructured/composite"

	"github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/resource"
)

// GetInput from the supplied request. Input is loaded into the supplied object.
func GetInput(req *v1beta1.RunFunctionRequest, into runtime.Object) error {
	return errors.Wrap(resource.AsObject(req.GetInput(), into), "cannot get Function input %T from %T, into, req")
}

// GetObservedCompositeResource from the supplied request.
func GetObservedCompositeResource(req *v1beta1.RunFunctionRequest) (*resource.Composite, error) {
	xr := &resource.Composite{
		Resource:          composite.New(),
		ConnectionDetails: req.GetObserved().GetComposite().GetConnectionDetails(),
	}
	err := resource.AsObject(req.GetObserved().GetComposite().GetResource(), xr.Resource)
	return xr, err
}

// GetObservedComposedResources from the supplied request.
func GetObservedComposedResources(req *v1beta1.RunFunctionRequest) (resource.ObservedComposedResources, error) {
	ocds := resource.ObservedComposedResources{}
	for name, r := range req.GetObserved().GetResources() {
		ocd := resource.ObservedComposed{Resource: composed.New(), ConnectionDetails: r.GetConnectionDetails()}
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
	err := resource.AsObject(req.GetDesired().GetComposite().GetResource(), xr.Resource)
	return xr, err
}

// GetDesiredComposedResources from the supplied request.
func GetDesiredComposedResources(req *v1beta1.RunFunctionRequest) (resource.DesiredComposedResources, error) {
	ocds := resource.DesiredComposedResources{}
	for name, r := range req.GetDesired().GetResources() {
		ocd := resource.DesiredComposed{Resource: composed.New()}
		if err := resource.AsObject(r.GetResource(), ocd.Resource); err != nil {
			return nil, err
		}
		ocds[resource.Name(name)] = ocd
	}
	return ocds, nil
}
