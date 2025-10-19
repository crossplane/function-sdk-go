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
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	v1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composed"
	"github.com/crossplane/function-sdk-go/resource/composite"
)

func TestGetObservedCompositeResource(t *testing.T) {
	type want struct {
		oxr *resource.Composite
		err error
	}

	cases := map[string]struct {
		reason string
		req    *v1.RunFunctionRequest
		want   want
	}{
		"NoObservedXR": {
			reason: "In the unlikely event the request has no observed XR we should return a usable, empty Composite.",
			req:    &v1.RunFunctionRequest{},
			want: want{
				oxr: &resource.Composite{
					Resource:          composite.New(),
					ConnectionDetails: resource.ConnectionDetails{},
				},
			},
		},
		"ObservedXR": {
			reason: "We should return the XR read from the request.",
			req: &v1.RunFunctionRequest{
				Observed: &v1.State{
					Composite: &v1.Resource{
						Resource: resource.MustStructJSON(`{
							"apiVersion": "test.crossplane.io/v1",
							"kind": "XR"
						}`),
						ConnectionDetails: map[string][]byte{
							"super": []byte("secret"),
						},
					},
				},
			},
			want: want{
				oxr: &resource.Composite{
					Resource: &composite.Unstructured{Unstructured: unstructured.Unstructured{
						Object: map[string]any{
							"apiVersion": "test.crossplane.io/v1",
							"kind":       "XR",
						},
					}},
					ConnectionDetails: resource.ConnectionDetails{
						"super": []byte("secret"),
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			oxr, err := GetObservedCompositeResource(tc.req)

			if diff := cmp.Diff(tc.want.oxr, oxr); diff != "" {
				t.Errorf("\n%s\nGetObservedCompositeResource(...): -want, +got:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("\n%s\nGetObservedCompositeResource(...): -want error, +got error:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestGetDesiredCompositeResource(t *testing.T) {
	type want struct {
		oxr *resource.Composite
		err error
	}

	cases := map[string]struct {
		reason string
		req    *v1.RunFunctionRequest
		want   want
	}{
		"NoDesiredXR": {
			reason: "If the request has no desired XR we should return a usable, empty Composite.",
			req:    &v1.RunFunctionRequest{},
			want: want{
				oxr: &resource.Composite{
					Resource:          composite.New(),
					ConnectionDetails: resource.ConnectionDetails{},
				},
			},
		},
		"DesiredXR": {
			reason: "We should return the XR read from the request.",
			req: &v1.RunFunctionRequest{
				Desired: &v1.State{
					Composite: &v1.Resource{
						Resource: resource.MustStructJSON(`{
							"apiVersion": "test.crossplane.io/v1",
							"kind": "XR"
						}`),
						ConnectionDetails: map[string][]byte{
							"super": []byte("secret"),
						},
					},
				},
			},
			want: want{
				oxr: &resource.Composite{
					Resource: &composite.Unstructured{Unstructured: unstructured.Unstructured{
						Object: map[string]any{
							"apiVersion": "test.crossplane.io/v1",
							"kind":       "XR",
						},
					}},
					ConnectionDetails: resource.ConnectionDetails{
						"super": []byte("secret"),
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			oxr, err := GetDesiredCompositeResource(tc.req)

			if diff := cmp.Diff(tc.want.oxr, oxr); diff != "" {
				t.Errorf("\n%s\nGetDesiredCompositeResource(...): -want, +got:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("\n%s\nGetDesiredCompositeResource(...): -want error, +got error:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestGetObservedComposedResources(t *testing.T) {
	type want struct {
		ocds map[resource.Name]resource.ObservedComposed
		err  error
	}

	cases := map[string]struct {
		reason string
		req    *v1.RunFunctionRequest
		want   want
	}{
		"NoObservedComposedResources": {
			reason: "If the request has no observed composed resources we should return an empty, non-nil map.",
			req:    &v1.RunFunctionRequest{},
			want: want{
				ocds: map[resource.Name]resource.ObservedComposed{},
			},
		},
		"ObservedComposedResources": {
			reason: "If the request has observed composed resources we should return them.",
			req: &v1.RunFunctionRequest{
				Observed: &v1.State{
					Resources: map[string]*v1.Resource{
						"observed-composed-resource": {
							Resource: resource.MustStructJSON(`{
								"apiVersion": "test.crossplane.io/v1",
								"kind": "Composed"
							}`),
							ConnectionDetails: map[string][]byte{
								"super": []byte("secret"),
							},
						},
					},
				},
			},
			want: want{
				ocds: map[resource.Name]resource.ObservedComposed{
					"observed-composed-resource": {
						Resource: &composed.Unstructured{Unstructured: unstructured.Unstructured{
							Object: map[string]any{
								"apiVersion": "test.crossplane.io/v1",
								"kind":       "Composed",
							},
						}},
						ConnectionDetails: resource.ConnectionDetails{
							"super": []byte("secret"),
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ocds, err := GetObservedComposedResources(tc.req)

			if diff := cmp.Diff(tc.want.ocds, ocds); diff != "" {
				t.Errorf("\n%s\nGetObservedComposedResources(...): -want, +got:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("\n%s\nGetObservedComposedResources(...): -want error, +got error:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestGetDesiredComposedResources(t *testing.T) {
	type want struct {
		dcds map[resource.Name]*resource.DesiredComposed
		err  error
	}

	cases := map[string]struct {
		reason string
		req    *v1.RunFunctionRequest
		want   want
	}{
		"NoDesiredComposedResources": {
			reason: "If the request has no desired composed resources we should return an empty, non-nil map.",
			req:    &v1.RunFunctionRequest{},
			want: want{
				dcds: map[resource.Name]*resource.DesiredComposed{},
			},
		},
		"DesiredComposedResources": {
			reason: "If the request has desired composed resources we should return them.",
			req: &v1.RunFunctionRequest{
				Desired: &v1.State{
					Resources: map[string]*v1.Resource{
						"desired-composed-resource": {
							Resource: resource.MustStructJSON(`{
								"apiVersion": "test.crossplane.io/v1",
								"kind": "Composed"
							}`),
							Ready: v1.Ready_READY_TRUE,
						},
					},
				},
			},
			want: want{
				dcds: map[resource.Name]*resource.DesiredComposed{
					"desired-composed-resource": {
						Resource: &composed.Unstructured{Unstructured: unstructured.Unstructured{
							Object: map[string]any{
								"apiVersion": "test.crossplane.io/v1",
								"kind":       "Composed",
							},
						}},
						Ready: resource.ReadyTrue,
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ocds, err := GetDesiredComposedResources(tc.req)

			if diff := cmp.Diff(tc.want.dcds, ocds); diff != "" {
				t.Errorf("\n%s\nGetDesiredComposedResources(...): -want, +got:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("\n%s\nGetDesiredComposedResources(...): -want error, +got error:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestGetRequiredResources(t *testing.T) {
	type want struct {
		resources map[string][]resource.Required
		err       error
	}

	cases := map[string]struct {
		reason string
		req    *v1.RunFunctionRequest
		want   want
	}{
		"NoRequiredResources": {
			reason: "If the request has no required resources we should return an empty, non-nil map.",
			req:    &v1.RunFunctionRequest{},
			want: want{
				resources: map[string][]resource.Required{},
			},
		},
		"RequiredResources": {
			reason: "If the request has required resources in the new field we should return them.",
			req: &v1.RunFunctionRequest{
				RequiredResources: map[string]*v1.Resources{
					"test-resources": {
						Items: []*v1.Resource{
							{
								Resource: resource.MustStructJSON(`{
									"apiVersion": "test.crossplane.io/v1",
									"kind": "TestResource",
									"metadata": {"name": "test"}
								}`),
							},
						},
					},
				},
			},
			want: want{
				resources: map[string][]resource.Required{
					"test-resources": {
						{
							Resource: &unstructured.Unstructured{
								Object: map[string]any{
									"apiVersion": "test.crossplane.io/v1",
									"kind":       "TestResource",
									"metadata":   map[string]any{"name": "test"},
								},
							},
						},
					},
				},
			},
		},
		"RequiredResourcesNamespaceScoped": {
			reason: "If the request has required namespace-scoped resources in the new field we should return them.",
			req: &v1.RunFunctionRequest{
				RequiredResources: map[string]*v1.Resources{
					"test-resources": {
						Items: []*v1.Resource{
							{
								Resource: resource.MustStructJSON(`{
									"apiVersion": "test.crossplane.io/v1",
									"kind": "TestResource",
									"metadata": {
										"name": "test", 
										"namespace": "test-namespace"
									}
								}`),
							},
						},
					},
				},
			},
			want: want{
				resources: map[string][]resource.Required{
					"test-resources": {
						{
							Resource: &unstructured.Unstructured{
								Object: map[string]any{
									"apiVersion": "test.crossplane.io/v1",
									"kind":       "TestResource",
									"metadata": map[string]any{
										"name":      "test",
										"namespace": "test-namespace",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			resources, err := GetRequiredResources(tc.req)

			if diff := cmp.Diff(tc.want.resources, resources); diff != "" {
				t.Errorf("\n%s\nGetRequiredResources(...): -want, +got:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("\n%s\nGetRequiredResources(...): -want error, +got error:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestGetExtraResources(t *testing.T) {
	type want struct {
		resources map[string][]resource.Required
		err       error
	}

	cases := map[string]struct {
		reason string
		req    *v1.RunFunctionRequest
		want   want
	}{
		"NoExtraResources": {
			reason: "If the request has no extra resources we should return an empty, non-nil map.",
			req:    &v1.RunFunctionRequest{},
			want: want{
				resources: map[string][]resource.Required{},
			},
		},
		"ExtraResources": {
			reason: "If the request has extra resources in the deprecated field we should return them.",
			req: &v1.RunFunctionRequest{
				ExtraResources: map[string]*v1.Resources{
					"test-resources": {
						Items: []*v1.Resource{
							{
								Resource: resource.MustStructJSON(`{
									"apiVersion": "test.crossplane.io/v1",
									"kind": "TestResource",
									"metadata": {"name": "test"}
								}`),
							},
						},
					},
				},
			},
			want: want{
				resources: map[string][]resource.Required{
					"test-resources": {
						{
							Resource: &unstructured.Unstructured{
								Object: map[string]any{
									"apiVersion": "test.crossplane.io/v1",
									"kind":       "TestResource",
									"metadata":   map[string]any{"name": "test"},
								},
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			resources, err := GetExtraResources(tc.req)

			if diff := cmp.Diff(tc.want.resources, resources); diff != "" {
				t.Errorf("\n%s\nGetExtraResources(...): -want, +got:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("\n%s\nGetExtraResources(...): -want error, +got error:\n%s", tc.reason, diff)
			}
		})
	}
}
