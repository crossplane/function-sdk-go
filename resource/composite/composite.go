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

// Package composite contains an unstructured composite resource (XR).
// This resource has getters and setters for common Kubernetes object metadata,
// as well as common composite resource fields like spec.claimRef. It also has
// generic fieldpath-based getters and setters to access arbitrary data.
package composite

import (
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/resource/unstructured/reference"
)

// New returns a new unstructured composite resource (XR).
func New() *Unstructured {
	return &Unstructured{unstructured.Unstructured{Object: make(map[string]any)}}
}

// An Unstructured composed resource (XR).
type Unstructured struct {
	unstructured.Unstructured
}

var (
	_ runtime.Object       = &Unstructured{}
	_ metav1.Object        = &Unstructured{}
	_ runtime.Unstructured = &Unstructured{}
	_ resource.Composite   = &Unstructured{}
)

// DeepCopy this composite resource.
func (xr *Unstructured) DeepCopy() *Unstructured {
	if xr == nil {
		return nil
	}
	out := new(Unstructured)
	*out = *xr
	out.Object = runtime.DeepCopyJSON(xr.Object)
	return out
}

// DeepCopyObject of this composite resource.
func (xr *Unstructured) DeepCopyObject() runtime.Object {
	return xr.DeepCopy()
}

// DeepCopyInto the supplied composite resource.
func (xr *Unstructured) DeepCopyInto(out *Unstructured) {
	clone := xr.DeepCopy()
	*out = *clone
}

// MarshalJSON for this composite resource.
func (xr *Unstructured) MarshalJSON() ([]byte, error) {
	return xr.Unstructured.MarshalJSON()
}

// GetCompositionSelector of this composite resource.
func (xr *Unstructured) GetCompositionSelector() *metav1.LabelSelector {
	out := &metav1.LabelSelector{}
	if err := fieldpath.Pave(xr.Object).GetValueInto("spec.compositionSelector", out); err != nil {
		return nil
	}
	return out
}

// SetCompositionSelector of this composite resource.
func (xr *Unstructured) SetCompositionSelector(sel *metav1.LabelSelector) {
	_ = fieldpath.Pave(xr.Object).SetValue("spec.compositionSelector", sel)
}

// GetCompositionReference of this composite resource.
func (xr *Unstructured) GetCompositionReference() *corev1.ObjectReference {
	out := &corev1.ObjectReference{}
	if err := fieldpath.Pave(xr.Object).GetValueInto("spec.compositionRef", out); err != nil {
		return nil
	}
	return out
}

// SetCompositionReference of this composite resource.
func (xr *Unstructured) SetCompositionReference(ref *corev1.ObjectReference) {
	_ = fieldpath.Pave(xr.Object).SetValue("spec.compositionRef", ref)
}

// GetCompositionRevisionReference of this composite resource.
func (xr *Unstructured) GetCompositionRevisionReference() *corev1.LocalObjectReference {
	out := &corev1.LocalObjectReference{}
	if err := fieldpath.Pave(xr.Object).GetValueInto("spec.compositionRevisionRef", out); err != nil {
		return nil
	}
	return out
}

// SetCompositionRevisionReference of this composite resource.
func (xr *Unstructured) SetCompositionRevisionReference(ref *corev1.LocalObjectReference) {
	_ = fieldpath.Pave(xr.Object).SetValue("spec.compositionRevisionRef", ref)
}

// GetCompositionRevisionSelector of this composite resource.
func (xr *Unstructured) GetCompositionRevisionSelector() *metav1.LabelSelector {
	out := &metav1.LabelSelector{}
	if err := fieldpath.Pave(xr.Object).GetValueInto("spec.compositionRevisionSelector", out); err != nil {
		return nil
	}
	return out
}

// SetCompositionRevisionSelector of this composite resource.
func (xr *Unstructured) SetCompositionRevisionSelector(sel *metav1.LabelSelector) {
	_ = fieldpath.Pave(xr.Object).SetValue("spec.compositionRevisionSelector", sel)
}

// SetCompositionUpdatePolicy of this composite resource.
func (xr *Unstructured) SetCompositionUpdatePolicy(p *xpv1.UpdatePolicy) {
	_ = fieldpath.Pave(xr.Object).SetValue("spec.compositionUpdatePolicy", p)
}

// GetCompositionUpdatePolicy of this composite resource.
func (xr *Unstructured) GetCompositionUpdatePolicy() *xpv1.UpdatePolicy {
	p, err := fieldpath.Pave(xr.Object).GetString("spec.compositionUpdatePolicy")
	if err != nil {
		return nil
	}
	out := xpv1.UpdatePolicy(p)
	return &out
}

// GetClaimReference of this composite resource.
func (xr *Unstructured) GetClaimReference() *reference.Claim {
	out := &reference.Claim{}
	if err := fieldpath.Pave(xr.Object).GetValueInto("spec.claimRef", out); err != nil {
		return nil
	}
	return out
}

// SetClaimReference of this composite resource.
func (xr *Unstructured) SetClaimReference(ref *reference.Claim) {
	_ = fieldpath.Pave(xr.Object).SetValue("spec.claimRef", ref)
}

// GetResourceReferences of this composite resource.
func (xr *Unstructured) GetResourceReferences() []corev1.ObjectReference {
	out := &[]corev1.ObjectReference{}
	_ = fieldpath.Pave(xr.Object).GetValueInto("spec.resourceRefs", out)
	return *out
}

// SetResourceReferences of this composite resource.
func (xr *Unstructured) SetResourceReferences(refs []corev1.ObjectReference) {
	empty := corev1.ObjectReference{}
	filtered := make([]corev1.ObjectReference, 0, len(refs))
	for _, ref := range refs {
		// TODO(negz): Ask muvaf to explain what this is working around. :)
		// TODO(muvaf): temporary workaround.
		if ref.String() == empty.String() {
			continue
		}
		filtered = append(filtered, ref)
	}
	_ = fieldpath.Pave(xr.Object).SetValue("spec.resourceRefs", filtered)
}

// GetWriteConnectionSecretToReference of this composite resource.
func (xr *Unstructured) GetWriteConnectionSecretToReference() *xpv1.SecretReference {
	out := &xpv1.SecretReference{}
	if err := fieldpath.Pave(xr.Object).GetValueInto("spec.writeConnectionSecretToRef", out); err != nil {
		return nil
	}
	return out
}

// SetWriteConnectionSecretToReference of this composite resource.
func (xr *Unstructured) SetWriteConnectionSecretToReference(ref *xpv1.SecretReference) {
	_ = fieldpath.Pave(xr.Object).SetValue("spec.writeConnectionSecretToRef", ref)
}

// GetPublishConnectionDetailsTo of this composite resource.
func (xr *Unstructured) GetPublishConnectionDetailsTo() *xpv1.PublishConnectionDetailsTo {
	out := &xpv1.PublishConnectionDetailsTo{}
	if err := fieldpath.Pave(xr.Object).GetValueInto("spec.publishConnectionDetailsTo", out); err != nil {
		return nil
	}
	return out
}

// SetPublishConnectionDetailsTo of this composite resource.
func (xr *Unstructured) SetPublishConnectionDetailsTo(ref *xpv1.PublishConnectionDetailsTo) {
	_ = fieldpath.Pave(xr.Object).SetValue("spec.publishConnectionDetailsTo", ref)
}

// GetCondition of this composite resource.
func (xr *Unstructured) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	conditioned := xpv1.ConditionedStatus{}
	// The path is directly `status` because conditions are inline.
	if err := fieldpath.Pave(xr.Object).GetValueInto("status", &conditioned); err != nil {
		return xpv1.Condition{}
	}
	return conditioned.GetCondition(ct)
}

// SetConditions of this composite resource.
func (xr *Unstructured) SetConditions(conditions ...xpv1.Condition) {
	conditioned := xpv1.ConditionedStatus{}
	// The path is directly `status` because conditions are inline.
	_ = fieldpath.Pave(xr.Object).GetValueInto("status", &conditioned)
	conditioned.SetConditions(conditions...)
	_ = fieldpath.Pave(xr.Object).SetValue("status.conditions", conditioned.Conditions)
}

// GetConnectionDetailsLastPublishedTime of this composite resource.
func (xr *Unstructured) GetConnectionDetailsLastPublishedTime() *metav1.Time {
	out := &metav1.Time{}
	if err := fieldpath.Pave(xr.Object).GetValueInto("status.connectionDetails.lastPublishedTime", out); err != nil {
		return nil
	}
	return out
}

// SetConnectionDetailsLastPublishedTime of this composite resource.
func (xr *Unstructured) SetConnectionDetailsLastPublishedTime(t *metav1.Time) {
	_ = fieldpath.Pave(xr.Object).SetValue("status.connectionDetails.lastPublishedTime", t)
}

// GetEnvironmentConfigReferences of this composite resource.
func (xr *Unstructured) GetEnvironmentConfigReferences() []corev1.ObjectReference {
	out := &[]corev1.ObjectReference{}
	_ = fieldpath.Pave(xr.Object).GetValueInto("spec.environmentConfigRefs", out)
	return *out
}

// SetEnvironmentConfigReferences of this composite resource.
func (xr *Unstructured) SetEnvironmentConfigReferences(refs []corev1.ObjectReference) {
	empty := corev1.ObjectReference{}
	filtered := make([]corev1.ObjectReference, 0, len(refs))
	for _, ref := range refs {
		// TODO(negz): Ask muvaf to explain what this is working around. :)
		// TODO(muvaf): temporary workaround.
		if ref.String() == empty.String() {
			continue
		}
		filtered = append(filtered, ref)
	}
	_ = fieldpath.Pave(xr.Object).SetValue("spec.environmentConfigRefs", filtered)
}

// GetValue of the supplied field path.
func (xr *Unstructured) GetValue(path string) (any, error) {
	return fieldpath.Pave(xr.Object).GetValue(path)
}

// GetValueInto the supplied type.
func (xr *Unstructured) GetValueInto(path string, out any) error {
	return fieldpath.Pave(xr.Object).GetValueInto(path, out)
}

// GetString value of the supplied field path.
func (xr *Unstructured) GetString(path string) (string, error) {
	return fieldpath.Pave(xr.Object).GetString(path)
}

// GetStringArray value of the supplied field path.
func (xr *Unstructured) GetStringArray(path string) ([]string, error) {
	return fieldpath.Pave(xr.Object).GetStringArray(path)
}

// GetStringObject value of the supplied field path.
func (xr *Unstructured) GetStringObject(path string) (map[string]string, error) {
	return fieldpath.Pave(xr.Object).GetStringObject(path)
}

// GetBool value of the supplied field path.
func (xr *Unstructured) GetBool(path string) (bool, error) {
	return fieldpath.Pave(xr.Object).GetBool(path)
}

// GetInteger value of the supplied field path.
func (xr *Unstructured) GetInteger(path string) (int64, error) {
	// This is a bit of a hack. Kubernetes JSON decoders will get us a
	// map[string]any where number values are int64, but protojson and structpb
	// will get us one where number values are float64.
	// https://pkg.go.dev/sigs.k8s.io/json#UnmarshalCaseSensitivePreserveInts
	p := fieldpath.Pave(xr.Object)

	// If we find an int64, return it.
	i64, err := p.GetInteger(path)
	if err == nil {
		return i64, nil
	}

	// If not, try return (and truncate) a float64.
	if f64, err := getNumber(p, path); err == nil {
		return int64(f64), nil
	}

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
func (xr *Unstructured) SetValue(path string, value any) error {
	return fieldpath.Pave(xr.Object).SetValue(path, value)
}

// SetString value at the supplied field path.
func (xr *Unstructured) SetString(path, value string) error {
	return xr.SetValue(path, value)
}

// SetBool value at the supplied field path.
func (xr *Unstructured) SetBool(path string, value bool) error {
	return xr.SetValue(path, value)
}

// SetInteger value at the supplied field path.
func (xr *Unstructured) SetInteger(path string, value int64) error {
	return xr.SetValue(path, value)
}

// SetObservedGeneration of this Composite resource.
func (xr *Unstructured) SetObservedGeneration(generation int64) {
	status := &xpv1.ObservedStatus{}
	_ = fieldpath.Pave(xr.Object).GetValueInto("status", status)
	status.SetObservedGeneration(generation)
	_ = fieldpath.Pave(xr.Object).SetValue("status.observedGeneration", status.ObservedGeneration)
}

// GetObservedGeneration of this Composite resource.
func (xr *Unstructured) GetObservedGeneration() int64 {
	status := &xpv1.ObservedStatus{}
	_ = fieldpath.Pave(xr.Object).GetValueInto("status", status)
	return status.GetObservedGeneration()
}
