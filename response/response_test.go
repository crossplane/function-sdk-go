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

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	v1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/resource"
)

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

func TestToDependencies(t *testing.T) {
	edge := &v1.Dependency{Resource: "subnet", DependsOn: &v1.Dependency_ComposedResource{ComposedResource: "vpc"}}

	cases := map[string]struct {
		reason string
		req    *v1.RunFunctionRequest
		want   *v1.Dependencies
	}{
		"Unset": {
			reason: "Unset dependencies mean no opinion, so they must stay unset rather than become an empty set.",
			req:    &v1.RunFunctionRequest{},
			want:   nil,
		},
		"Empty": {
			reason: "An empty set means drop every constraint, and must be carried as such.",
			req:    &v1.RunFunctionRequest{Dependencies: &v1.Dependencies{}},
			want:   &v1.Dependencies{},
		},
		"Carried": {
			reason: "A function should keep the dependencies earlier functions declared.",
			req:    &v1.RunFunctionRequest{Dependencies: &v1.Dependencies{Items: []*v1.Dependency{edge}}},
			want:   &v1.Dependencies{Items: []*v1.Dependency{edge}},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := To(tc.req, DefaultTTL).GetDependencies()
			if diff := cmp.Diff(tc.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("%s\nTo(...).Dependencies: -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestToDoesNotShareDependencies(t *testing.T) {
	req := &v1.RunFunctionRequest{Dependencies: &v1.Dependencies{}}
	rsp := To(req, DefaultTTL)
	AddDependency(rsp, "subnet", "vpc")

	if n := len(req.GetDependencies().GetItems()); n != 0 {
		t.Errorf("adding a dependency to the response added %d to the request", n)
	}
}

func TestAddDependency(t *testing.T) {
	type args struct {
		rsp       *v1.RunFunctionResponse
		r         resource.Name
		dependsOn resource.Name
		o         []DependencyOption
	}
	cases := map[string]struct {
		reason string
		args   args
		want   *v1.RunFunctionResponse
	}{
		"FirstDependency": {
			reason: "Adding to a response with no dependencies should create the set.",
			args:   args{rsp: &v1.RunFunctionResponse{}, r: "subnet", dependsOn: "vpc"},
			want: &v1.RunFunctionResponse{Dependencies: &v1.Dependencies{Items: []*v1.Dependency{
				{Resource: "subnet", DependsOn: &v1.Dependency_ComposedResource{ComposedResource: "vpc"}},
			}}},
		},
		"Appends": {
			reason: "Adding should keep the dependencies already there.",
			args: args{
				rsp: &v1.RunFunctionResponse{Dependencies: &v1.Dependencies{Items: []*v1.Dependency{
					{Resource: "subnet", DependsOn: &v1.Dependency_ComposedResource{ComposedResource: "vpc"}},
				}}},
				r:         "instance",
				dependsOn: "subnet",
			},
			want: &v1.RunFunctionResponse{Dependencies: &v1.Dependencies{Items: []*v1.Dependency{
				{Resource: "subnet", DependsOn: &v1.Dependency_ComposedResource{ComposedResource: "vpc"}},
				{Resource: "instance", DependsOn: &v1.Dependency_ComposedResource{ComposedResource: "subnet"}},
			}}},
		},
		"CreateBeforeDestroy": {
			reason: "WithCreateBeforeDestroy should set the lifecycle.",
			args: args{
				rsp: &v1.RunFunctionResponse{}, r: "new-db", dependsOn: "old-db",
				o: []DependencyOption{WithCreateBeforeDestroy()},
			},
			want: &v1.RunFunctionResponse{Dependencies: &v1.Dependencies{Items: []*v1.Dependency{
				{
					Resource:  "new-db",
					DependsOn: &v1.Dependency_ComposedResource{ComposedResource: "old-db"},
					Lifecycle: v1.DependencyLifecycle_DEPENDENCY_LIFECYCLE_CREATE_BEFORE_DESTROY,
				},
			}}},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			AddDependency(tc.args.rsp, tc.args.r, tc.args.dependsOn, tc.args.o...)
			if diff := cmp.Diff(tc.want, tc.args.rsp, protocmp.Transform()); diff != "" {
				t.Errorf("%s\nAddDependency(...): -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestAddRequiredResourceDependency(t *testing.T) {
	rsp := &v1.RunFunctionResponse{}
	AddRequiredResourceDependency(rsp, "app-config", &v1.RequiredResourceDependency{RequirementName: "dbs"})

	want := &v1.RunFunctionResponse{Dependencies: &v1.Dependencies{Items: []*v1.Dependency{
		{
			Resource:  "app-config",
			DependsOn: &v1.Dependency_RequiredResource{RequiredResource: &v1.RequiredResourceDependency{RequirementName: "dbs"}},
		},
	}}}
	if diff := cmp.Diff(want, rsp, protocmp.Transform()); diff != "" {
		t.Errorf("AddRequiredResourceDependency(...): -want, +got:\n%s", diff)
	}
}

func TestClearDependencies(t *testing.T) {
	rsp := &v1.RunFunctionResponse{Dependencies: &v1.Dependencies{Items: []*v1.Dependency{
		{Resource: "subnet", DependsOn: &v1.Dependency_ComposedResource{ComposedResource: "vpc"}},
	}}}
	ClearDependencies(rsp)

	// Empty, not unset: unset would carry the dropped dependencies forward.
	if rsp.GetDependencies() == nil || len(rsp.GetDependencies().GetItems()) != 0 {
		t.Errorf("ClearDependencies(...): want an empty set, got %v", rsp.GetDependencies())
	}
}
