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
	"github.com/go-json-experiment/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/function-sdk-go/errors"
)

// Scheme used to determine the type of any runtime.Object passed to From.
var Scheme *runtime.Scheme

func init() {
	Scheme = runtime.NewScheme()
}

// New returns a new unstructured composed resource.
func New() *Unstructured {
	return &Unstructured{unstructured.Unstructured{Object: make(map[string]any)}}
}

// From creates a new unstructured composed resource from the supplied object.
func From(o runtime.Object) (*Unstructured, error) {
	// If the supplied object is already unstructured content, avoid a JSON
	// round trip and use it.
	if u, ok := o.(interface{ UnstructuredContent() map[string]any }); ok {
		return &Unstructured{unstructured.Unstructured{Object: u.UnstructuredContent()}}, nil
	}

	// Set the object's GVK from our scheme.
	gvks, _, err := Scheme.ObjectKinds(o)
	if err != nil {
		return nil, errors.Wrap(err, "did you add it to composed.Scheme?")
	}
	// There should almost never be more than one GVK for a type.
	for _, gvk := range gvks {
		o.GetObjectKind().SetGroupVersionKind(gvk)
	}

	// Round-trip the supplied object through JSON to convert it. We use the
	// go-json-experiment package for this because it honors the omitempty field
	// for non-pointer struct fields.
	//
	// At the time of writing many Crossplane structs contain fields that have
	// the omitempty struct tag, but non-pointer struct values. pkg/json does
	// not omit these fields. It instead includes them as empty JSON objects.
	// Crossplane will interpret this as part of a server-side apply fully
	// specified intent and assume the function actually has opinion about the
	// field when it doesn't. We should make these fields pointers, but it's
	// easier and safer in the meantime to work around it here.
	//
	// https://github.com/go-json-experiment/json#behavior-changes
	j, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	obj := make(map[string]any)
	if err := json.Unmarshal(j, &obj); err != nil {
		return nil, err
	}

	// Unfortunately we still need to cleanup some object metadata.
	cleanupMetadata(obj)

	return &Unstructured{unstructured.Unstructured{Object: obj}}, nil
}

func cleanupMetadata(obj map[string]any) {
	m, ok := obj["metadata"]
	if !ok {
		// If there's no metadata there's nothing to do.
		return
	}

	mo, ok := m.(map[string]any)
	if !ok {
		// If metadata isn't an object there's nothing to do.
		return
	}

	// The ObjectMeta struct that all Kubernetes types include has a non-nil
	// integer Generation field with the omitempty tag. Regular pkg/json removes
	// this, but go-json-experiment does not (it would need the new omitzero
	// tag). So, we clean it up manually. No function should ever be setting it.
	delete(mo, "generation")

	// If metadata has no fields, delete it. This prevents us from serializing
	// metadata: {}, which SSA would interpret as "make metadata empty".
	if len(mo) == 0 {
		delete(obj, "metadata")
	}
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

// SetObservedGeneration of this composite resource claim.
func (cd *Unstructured) SetObservedGeneration(generation int64) {
	status := &xpv1.ObservedStatus{}
	_ = fieldpath.Pave(cd.Object).GetValueInto("status", status)
	status.SetObservedGeneration(generation)
	_ = fieldpath.Pave(cd.Object).SetValue("status.observedGeneration", status.ObservedGeneration)
}

// GetObservedGeneration of this composite resource claim.
func (cd *Unstructured) GetObservedGeneration() int64 {
	status := &xpv1.ObservedStatus{}
	_ = fieldpath.Pave(cd.Object).GetValueInto("status", status)
	return status.GetObservedGeneration()
}
