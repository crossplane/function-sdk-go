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

// Package resource contains utilities to convert protobuf representations of
// Crossplane resources to unstructured Go types, often with convenient getters
// and setters.
package resource

import (
	"github.com/go-json-experiment/json"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/crossplane/function-sdk-go/errors"
	"github.com/crossplane/function-sdk-go/resource/composed"
	"github.com/crossplane/function-sdk-go/resource/composite"
)

// ConnectionDetails created or updated during an operation on an external
// resource, for example usernames, passwords, endpoints, ports, etc.
type ConnectionDetails map[string][]byte

// A Composite resource - aka an XR.
type Composite struct {
	Resource          *composite.Unstructured
	ConnectionDetails ConnectionDetails
}

// A Name uniquely identifies a composed resource within a Composition Function
// pipeline. It's not the resource's metadata.name.
type Name string

// DesiredComposed reflects the desired state of a composed resource.
type DesiredComposed struct {
	Resource *composed.Unstructured

	Ready Ready
}

// Extra is a resource requested by a Function.
type Extra struct {
	Resource *unstructured.Unstructured
}

// Credential is a secret requested by a Function
type Credential struct {
	Data *unstructured.Unstructured
}

// Ready indicates whether a composed resource should be considered ready.
type Ready string

// Composed resource readiness.
const (
	ReadyUnspecified Ready = "Unspecified"
	ReadyTrue        Ready = "True"
	ReadyFalse       Ready = "False"
)

// NewDesiredComposed returns a new, empty desired composed resource.
func NewDesiredComposed() *DesiredComposed {
	return &DesiredComposed{Resource: composed.New()}
}

// ObservedComposed reflects the observed state of a composed resource.
type ObservedComposed struct {
	Resource          *composed.Unstructured
	ConnectionDetails ConnectionDetails
}

// AsObject gets the supplied Kubernetes object from the supplied struct.
func AsObject(s *structpb.Struct, o runtime.Object) error {
	// We try to avoid a JSON round-trip if o is backed by unstructured data.
	// Any type that is or embeds *unstructured.Unstructured has this method.
	if u, ok := o.(interface{ SetUnstructuredContent(map[string]any) }); ok {
		u.SetUnstructuredContent(s.AsMap())
		return nil
	}

	b, err := protojson.Marshal(s)
	if err != nil {
		return errors.Wrapf(err, "cannot marshal %T to JSON", s)
	}
	return errors.Wrapf(json.Unmarshal(b, o, json.RejectUnknownMembers(true)), "cannot unmarshal JSON from %T into %T", s, o)
}

// AsStruct gets the supplied struct from the supplied Kubernetes object.
func AsStruct(o runtime.Object) (*structpb.Struct, error) {
	// We try to avoid a JSON round-trip if o is backed by unstructured data.
	// Any type that is or embeds *unstructured.Unstructured has this method.
	if u, ok := o.(interface{ UnstructuredContent() map[string]any }); ok {
		s, err := structpb.NewStruct(u.UnstructuredContent())
		return s, errors.Wrapf(err, "cannot create new Struct from %T", u)
	}

	b, err := json.Marshal(o)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot marshal %T to JSON", o)
	}
	s := &structpb.Struct{}
	return s, errors.Wrapf(protojson.Unmarshal(b, s), "cannot unmarshal JSON from %T into %T", o, s)
}

// MustStructObject is intended only for use in tests. It returns the supplied
// object as a struct. It panics if it can't.
func MustStructObject(o runtime.Object) *structpb.Struct {
	s, err := AsStruct(o)
	if err != nil {
		panic(err)
	}
	return s
}

// MustStructJSON is intended only for use in tests. It returns the supplied
// JSON string as a struct. It panics if it can't.
func MustStructJSON(j string) *structpb.Struct {
	s := &structpb.Struct{}
	if err := protojson.Unmarshal([]byte(j), s); err != nil {
		panic(err)
	}
	return s
}
