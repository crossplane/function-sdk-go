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

package function

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/crossplane/function-sdk-go/errors"
	v1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composed"
	"github.com/crossplane/function-sdk-go/response"
)

var _ v1beta1.FunctionRunnerServiceServer = &BetaServer{}

var req = &v1.RunFunctionRequest{
	Observed: &v1.State{
		Composite: &v1.Resource{
			Resource: resource.MustStructJSON(`{"spec":{"widgets":9001}}`),
		},
	},
}

func Example() {
	// Create a response to the request passed to your RunFunction method.
	rsp := response.To(req, response.DefaultTTL)

	// Get the observed composite resource (XR) from the request.
	oxr, _ := request.GetObservedCompositeResource(req)

	// Read the desired number of widgets from our observed XR.
	widgets, _ := oxr.Resource.GetInteger("spec.widgets")

	// Get any existing desired composed resources from the request.
	// Desired composed resources would exist if a previous Function in the
	// pipeline added them.
	desired, _ := request.GetDesiredComposedResources(req)

	// Create a desired composed resource using unstructured data.
	desired["new"] = &resource.DesiredComposed{Resource: composed.New()}
	desired["new"].Resource.SetAPIVersion("example.org/v1")
	desired["new"].Resource.SetKind("CoolResource")

	// Set the desired composed resource's widgets to the value extracted from
	// the observed XR.
	desired["new"].Resource.SetInteger("spec.widgets", widgets)

	// Create a desired composed resource using structured data.
	// db, _ := composed.From(&v1.Instance{})
	// desired["database"] = &resource.DesiredComposed{Resource: db}

	// Add a label to our new desired resource, and any other.
	for _, r := range desired {
		r.Resource.SetLabels(map[string]string{"coolness": "high"})
	}

	// Set our updated desired composed resource in the response we'll return.
	_ = response.SetDesiredComposedResources(rsp, desired)

	j, _ := protojson.Marshal(rsp)
	fmt.Println(string(j))

	// Output:
	// {"meta":{"ttl":"60s"},"desired":{"resources":{"new":{"resource":{"apiVersion":"example.org/v1","kind":"CoolResource","metadata":{"labels":{"coolness":"high"}},"spec":{"widgets":9001}}}}}}
}

func TestBetaServer(t *testing.T) {
	type args struct {
		ctx context.Context
		req *v1beta1.RunFunctionRequest
	}
	type want struct {
		rsp *v1beta1.RunFunctionResponse
		err error
	}

	cases := map[string]struct {
		reason  string
		wrapped v1.FunctionRunnerServiceServer
		args    args
		want    want
	}{
		"RunFunctionError": {
			reason:  "We should return any error the wrapped server encounters",
			wrapped: &MockFunctionServer{err: errors.New("boom")},
			args: args{
				req: &v1beta1.RunFunctionRequest{
					Meta: &v1beta1.RequestMeta{
						Tag: "hi",
					},
				},
			},
			want: want{
				err: cmpopts.AnyError,
			},
		},
		"Success": {
			reason: "We should return the response the wrapped server returns",
			wrapped: &MockFunctionServer{
				rsp: &v1.RunFunctionResponse{
					Meta: &v1.ResponseMeta{
						Tag: "hello",
					},
				},
			},
			args: args{
				req: &v1beta1.RunFunctionRequest{
					Meta: &v1beta1.RequestMeta{
						Tag: "hi",
					},
				},
			},
			want: want{
				rsp: &v1beta1.RunFunctionResponse{
					Meta: &v1beta1.ResponseMeta{
						Tag: "hello",
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			s := ServeBeta(tc.wrapped)
			rsp, err := s.RunFunction(tc.args.ctx, tc.args.req)

			if diff := cmp.Diff(tc.want.rsp, rsp, protocmp.Transform()); diff != "" {
				t.Errorf("\n%s\ns.RunFunction(...): -want rsp, +got rsp:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ns.RunFunction(...): -want err, +got err:\n%s", tc.reason, diff)
			}
		})
	}

}

type MockFunctionServer struct {
	v1.UnimplementedFunctionRunnerServiceServer

	rsp *v1.RunFunctionResponse
	err error
}

func (s *MockFunctionServer) RunFunction(context.Context, *v1.RunFunctionRequest) (*v1.RunFunctionResponse, error) {
	return s.rsp, s.err
}
