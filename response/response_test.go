/*
Copyright 2025 The Crossplane Authors.

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

package response

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	v1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/resource"
)

func TestToClonesRequestState(t *testing.T) {
	req := &v1.RunFunctionRequest{
		Meta: &v1.RequestMeta{Tag: "original"},
		Desired: &v1.State{
			Resources: map[string]*v1.Resource{
				"existing": {
					Resource: resource.MustStructJSON(`{
						"apiVersion": "example.org/v1",
						"kind": "Widget"
					}`),
				},
			},
		},
		Context: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"existing": structpb.NewStringValue("value"),
			},
		},
	}

	rsp := To(req, time.Minute)

	SetContextKey(rsp, "new", structpb.NewStringValue("response-only"))
	rsp.Desired.Resources["response-only"] = &v1.Resource{
		Resource: resource.MustStructJSON(`{
			"apiVersion": "example.org/v1",
			"kind": "ExtraWidget"
		}`),
	}

	if got, ok := req.GetContext().GetFields()["new"]; ok {
		t.Fatalf("To(...) aliased request context: unexpected field %q", got.GetStringValue())
	}

	if _, ok := req.GetDesired().GetResources()["response-only"]; ok {
		t.Fatal("To(...) aliased desired resources from the request")
	}
}

func TestSetDesiredResources(t *testing.T) {
	type args struct {
		rsp *v1.RunFunctionResponse
		drs map[resource.Name]*unstructured.Unstructured
	}
	type want struct {
		rsp *v1.RunFunctionResponse
		err error
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"Success": {
			args: args{
				rsp: &v1.RunFunctionResponse{},
				drs: map[resource.Name]*unstructured.Unstructured{
					"Cool": MustUnstructJSON(`{
						"apiVersion": "example.org/v1",
						"kind": "Test",
						"metadata": {
							"name": "cool"
						},
						"spec" : {
							"cool": true
						}
					}`),
				},
			},
			want: want{
				rsp: &v1.RunFunctionResponse{
					Desired: &v1.State{
						Resources: map[string]*v1.Resource{
							"Cool": {
								Resource: resource.MustStructJSON(`{
									"apiVersion": "example.org/v1",
									"kind": "Test",
									"metadata": {
										"name": "cool"
									},
									"spec" : {
										"cool": true
									}
								}`),
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := SetDesiredResources(tc.args.rsp, tc.args.drs)

			if diff := cmp.Diff(tc.want.rsp, tc.args.rsp, protocmp.Transform()); diff != "" {
				t.Errorf("SetDesiredResources(...): -want rsp, +got rsp:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("SetDesiredResources(...): -want err, +got err:\n%s", diff)
			}
		})
	}
}

func TestOutput(t *testing.T) {
	type out struct {
		Cool string `json:"cool"`
	}

	type args struct {
		rsp    *v1.RunFunctionResponse
		output any
	}
	type want struct {
		rsp *v1.RunFunctionResponse
		err error
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"Unmarshalable": {
			args: args{
				rsp:    &v1.RunFunctionResponse{},
				output: make(chan<- bool),
			},
			want: want{
				rsp: &v1.RunFunctionResponse{},
				err: cmpopts.AnyError,
			},
		},
		"Success": {
			args: args{
				rsp:    &v1.RunFunctionResponse{},
				output: &out{Cool: "very"},
			},
			want: want{
				rsp: &v1.RunFunctionResponse{
					Output: resource.MustStructJSON(`{
						"cool": "very"
					}`),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := SetOutput(tc.args.rsp, tc.args.output)

			if diff := cmp.Diff(tc.want.rsp, tc.args.rsp, protocmp.Transform()); diff != "" {
				t.Errorf("SetDesiredResources(...): -want rsp, +got rsp:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("SetDesiredResources(...): -want err, +got err:\n%s", diff)
			}
		})
	}
}

func TestRequireSchema(t *testing.T) {
	type args struct {
		rsp        *v1.RunFunctionResponse
		name       string
		apiVersion string
		kind       string
	}
	type want struct {
		rsp *v1.RunFunctionResponse
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"NewRequirement": {
			args: args{
				rsp:        &v1.RunFunctionResponse{},
				name:       "xr-schema",
				apiVersion: "example.org/v1",
				kind:       "MyResource",
			},
			want: want{
				rsp: &v1.RunFunctionResponse{
					Requirements: &v1.Requirements{
						Schemas: map[string]*v1.SchemaSelector{
							"xr-schema": {
								ApiVersion: "example.org/v1",
								Kind:       "MyResource",
							},
						},
					},
				},
			},
		},
		"ExistingRequirements": {
			args: args{
				rsp: &v1.RunFunctionResponse{
					Requirements: &v1.Requirements{
						Resources: map[string]*v1.ResourceSelector{
							"existing": {ApiVersion: "v1", Kind: "ConfigMap"},
						},
					},
				},
				name:       "xr-schema",
				apiVersion: "example.org/v1",
				kind:       "MyResource",
			},
			want: want{
				rsp: &v1.RunFunctionResponse{
					Requirements: &v1.Requirements{
						Resources: map[string]*v1.ResourceSelector{
							"existing": {ApiVersion: "v1", Kind: "ConfigMap"},
						},
						Schemas: map[string]*v1.SchemaSelector{
							"xr-schema": {
								ApiVersion: "example.org/v1",
								Kind:       "MyResource",
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			RequireSchema(tc.args.rsp, tc.args.name, tc.args.apiVersion, tc.args.kind)

			if diff := cmp.Diff(tc.want.rsp, tc.args.rsp, protocmp.Transform()); diff != "" {
				t.Errorf("RequireSchema(...): -want rsp, +got rsp:\n%s", diff)
			}
		})
	}
}

func MustUnstructJSON(j string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	if err := json.Unmarshal([]byte(j), u); err != nil {
		panic(err)
	}
	return u
}
