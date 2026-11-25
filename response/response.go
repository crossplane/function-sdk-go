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
	"google.golang.org/protobuf/proto"
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
// desired state, context and dependencies from the request.
//
// Copying dependencies means a function that adds one keeps the ones earlier
// functions declared. They're copied only when the request has them: unset
// dependencies mean "no opinion", which Crossplane reads as "carry mine
// forward", while an empty set means "drop every constraint".
func To(req *v1.RunFunctionRequest, ttl time.Duration) *v1.RunFunctionResponse {
	return &v1.RunFunctionResponse{
		Meta: &v1.ResponseMeta{
			Tag: req.GetMeta().GetTag(),
			Ttl: durationpb.New(ttl),
		},
		Desired:      req.GetDesired(),
		Context:      req.GetContext(),
		Dependencies: dependenciesOf(req),
	}
}

// dependenciesOf copies the request's dependencies, so adding one to the
// response doesn't also add it to the request.
func dependenciesOf(req *v1.RunFunctionRequest) *v1.Dependencies {
	if req.GetDependencies() == nil {
		return nil
	}
	d, _ := proto.Clone(req.GetDependencies()).(*v1.Dependencies)
	return d
}

// A DependencyOption configures a dependency added by AddDependency.
type DependencyOption func(d *v1.Dependency)

// WithCreateBeforeDestroy lets a resource be created without waiting for what
// it depends on to be deleted. Use it for a replacement that must exist before
// its predecessor is torn down.
func WithCreateBeforeDestroy() DependencyOption {
	return func(d *v1.Dependency) {
		d.Lifecycle = v1.DependencyLifecycle_DEPENDENCY_LIFECYCLE_CREATE_BEFORE_DESTROY
	}
}

// AddDependency declares that one composed resource depends on another.
//
// By default ordering is symmetric: the resource is created only once what it
// depends on is ready, and what it depends on is deleted only once the
// resource is gone. Dependencies express ordering only; they don't move any
// data between resources.
//
// A function must return the full set of dependencies it wants. To copies
// forward the ones the request carried, so add to a response it created.
//
// Only a Crossplane that advertises CAPABILITY_DEPENDENCIES honors
// dependencies. Use request.HasCapability to check before relying on them.
func AddDependency(rsp *v1.RunFunctionResponse, r, dependsOn resource.Name, o ...DependencyOption) {
	d := &v1.Dependency{
		Resource:  string(r),
		DependsOn: &v1.Dependency_ComposedResource{ComposedResource: string(dependsOn)},
	}
	for _, fn := range o {
		fn(d)
	}
	addDependency(rsp, d)
}

// AddRequiredResourceDependency declares that a composed resource depends on
// a resource the function requires but doesn't compose. Set the dependency's
// name, and namespace for a namespaced resource, to depend on one of the
// resources the requirement matched; leave them unset to wait for all of them.
//
// Crossplane never deletes a resource it didn't compose, so this orders only
// creation and updates.
func AddRequiredResourceDependency(rsp *v1.RunFunctionResponse, r resource.Name, dependsOn *v1.RequiredResourceDependency) {
	addDependency(rsp, &v1.Dependency{
		Resource:  string(r),
		DependsOn: &v1.Dependency_RequiredResource{RequiredResource: dependsOn},
	})
}

// ClearDependencies declares that no composed resources should be ordered. It
// drops the dependencies earlier functions declared, which To copied forward.
// That's different from leaving dependencies unset, which carries them
// forward.
func ClearDependencies(rsp *v1.RunFunctionResponse) {
	rsp.Dependencies = &v1.Dependencies{}
}

func addDependency(rsp *v1.RunFunctionResponse, d *v1.Dependency) {
	if rsp.GetDependencies() == nil {
		rsp.Dependencies = &v1.Dependencies{}
	}
	rsp.Dependencies.Items = append(rsp.Dependencies.Items, d)
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

// RequireSchema adds a schema requirement to the response. This tells
// Crossplane to fetch the OpenAPI schema for the specified resource kind and
// include it in the next request's required_schemas field. Use
// request.GetRequiredSchema to retrieve the resolved schema.
//
// For CRDs, Crossplane returns the spec.versions[].schema.openAPIV3Schema field.
// If Crossplane cannot find a schema for the requested kind, the schema will be
// empty (GetRequiredSchema will return nil with ok true).
func RequireSchema(rsp *v1.RunFunctionResponse, name, apiVersion, kind string) {
	if rsp.GetRequirements() == nil {
		rsp.Requirements = &v1.Requirements{}
	}
	if rsp.Requirements.Schemas == nil {
		rsp.Requirements.Schemas = make(map[string]*v1.SchemaSelector)
	}
	rsp.Requirements.Schemas[name] = &v1.SchemaSelector{
		ApiVersion: apiVersion,
		Kind:       kind,
	}
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
