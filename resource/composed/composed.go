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

// Package composed contains an unstructured composed resource.
package composed

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/function-sdk-go/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// New returns a new unstructured composed resource.
func New() *Unstructured {
	return &Unstructured{unstructured.Unstructured{Object: make(map[string]any)}}
}

// From creates a new unstructured composed resource from the supplied object.
func From(o runtime.Object) (*Unstructured, error) {
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(o)
	if err != nil {
		return nil, err
	}
	return &Unstructured{unstructured.Unstructured{Object: obj}}, nil
}

// An Unstructured composed resource.
type Unstructured struct {
	unstructured.Unstructured
}

var _ runtime.Object = &Unstructured{}
var _ metav1.Object = &Unstructured{}
var _ runtime.Unstructured = &Unstructured{}
var _ resource.Composed = &Unstructured{}

// DeepCopy this composed resource.
func (cd *Unstructured) DeepCopy() *Unstructured {
	if cd == nil {
		return nil
	}
	out := new(Unstructured)
	*out = *cd
	out.Object = runtime.DeepCopyJSON(cd.Object)
	return out
}

// DeepCopyObject of this composed resource.
func (cd *Unstructured) DeepCopyObject() runtime.Object {
	return cd.DeepCopy()
}

// DeepCopyInto the supplied composed resource.
func (cd *Unstructured) DeepCopyInto(out *Unstructured) {
	clone := cd.DeepCopy()
	*out = *clone
}

// MarshalJSON for this composed resource.
func (cd *Unstructured) MarshalJSON() ([]byte, error) {
	return cd.Unstructured.MarshalJSON()
}

// GetCondition of this Composed resource.
func (cd *Unstructured) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	conditioned := xpv1.ConditionedStatus{}
	// The path is directly `status` because conditions are inline.
	if err := fieldpath.Pave(cd.Object).GetValueInto("status", &conditioned); err != nil {
		return xpv1.Condition{}
	}
	return conditioned.GetCondition(ct)
}

// SetConditions of this Composed resource.
func (cd *Unstructured) SetConditions(c ...xpv1.Condition) {
	conditioned := xpv1.ConditionedStatus{}
	// The path is directly `status` because conditions are inline.
	_ = fieldpath.Pave(cd.Object).GetValueInto("status", &conditioned)
	conditioned.SetConditions(c...)
	_ = fieldpath.Pave(cd.Object).SetValue("status.conditions", conditioned.Conditions)
}

// GetWriteConnectionSecretToReference of this Composed resource.
func (cd *Unstructured) GetWriteConnectionSecretToReference() *xpv1.SecretReference {
	out := &xpv1.SecretReference{}
	if err := fieldpath.Pave(cd.Object).GetValueInto("spec.writeConnectionSecretToRef", out); err != nil {
		return nil
	}
	return out
}

// SetWriteConnectionSecretToReference of this Composed resource.
func (cd *Unstructured) SetWriteConnectionSecretToReference(r *xpv1.SecretReference) {
	_ = fieldpath.Pave(cd.Object).SetValue("spec.writeConnectionSecretToRef", r)
}

// GetPublishConnectionDetailsTo of this Composed resource.
func (cd *Unstructured) GetPublishConnectionDetailsTo() *xpv1.PublishConnectionDetailsTo {
	out := &xpv1.PublishConnectionDetailsTo{}
	if err := fieldpath.Pave(cd.Object).GetValueInto("spec.publishConnectionDetailsTo", out); err != nil {
		return nil
	}
	return out
}

// SetPublishConnectionDetailsTo of this Composed resource.
func (cd *Unstructured) SetPublishConnectionDetailsTo(ref *xpv1.PublishConnectionDetailsTo) {
	_ = fieldpath.Pave(cd.Object).SetValue("spec.publishConnectionDetailsTo", ref)
}

// GetValue of the supplied field path.
func (cd *Unstructured) GetValue(path string) (any, error) {
	return fieldpath.Pave(cd.Object).GetValue(path)
}

// GetValueInto the supplied type.
func (cd *Unstructured) GetValueInto(path string, out any) error {
	return fieldpath.Pave(cd.Object).GetValueInto(path, out)
}

// GetString value of the supplied field path.
func (cd *Unstructured) GetString(path string) (string, error) {
	return fieldpath.Pave(cd.Object).GetString(path)
}

// GetStringArray value of the supplied field path.
func (cd *Unstructured) GetStringArray(path string) ([]string, error) {
	return fieldpath.Pave(cd.Object).GetStringArray(path)
}

// GetStringObject value of the supplied field path.
func (cd *Unstructured) GetStringObject(path string) (map[string]string, error) {
	return fieldpath.Pave(cd.Object).GetStringObject(path)
}

// GetBool value of the supplied field path.
func (cd *Unstructured) GetBool(path string) (bool, error) {
	return fieldpath.Pave(cd.Object).GetBool(path)
}

// GetInteger value of the supplied field path.
func (cd *Unstructured) GetInteger(path string) (int64, error) {
	// This is a bit of a hack. Kubernetes JSON decoders will get us a
	// map[string]any where number values are int64, but protojson and structpb
	// will get us one where number values are float64.
	// https://pkg.go.dev/sigs.k8s.io/json#UnmarshalCaseSensitivePreserveInts
	p := fieldpath.Pave(cd.Object)

	// If we find an int64, return it.
	i64, err := p.GetInteger(path)
	if err == nil {
		return i64, nil

	}

	// If not, try return (and truncate) a float64.
	if f64, err := getNumber(p, path); err == nil {
		return int64(f64), nil
	}

	// If both fail, return our original error.
	return 0, err
}

func getNumber(p *fieldpath.Paved, path string) (float64, error) {
	v, err := p.GetValue(path)
	if err != nil {
		return 0, err
	}

	f, ok := v.(float64)
	if !ok {
		return 0, errors.Errorf("%s: not a (float64) number", path)
	}
	return f, nil
}

// SetValue at the supplied field path.
func (cd *Unstructured) SetValue(path string, value any) error {
	return fieldpath.Pave(cd.Object).SetValue(path, value)
}

// SetString value at the supplied field path.
func (cd *Unstructured) SetString(path, value string) error {
	return cd.SetValue(path, value)
}

// SetBool value at the supplied field path.
func (cd *Unstructured) SetBool(path string, value bool) error {
	return cd.SetValue(path, value)
}

// SetInteger value at the supplied field path.
func (cd *Unstructured) SetInteger(path string, value int64) error {
	return cd.SetValue(path, value)
}
