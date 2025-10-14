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
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	if err := response.SetDesiredComposedResources(rsp, desired); err != nil {
		// You can set a custom status condition on the claim. This allows you to
		// communicate with the user. See the link below for status condition
		// guidance.
		// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
		response.ConditionFalse(rsp, "FunctionSuccess", "InternalError").
			WithMessage("Something went wrong.").
			TargetCompositeAndClaim()

		// You can emit an event regarding the claim. This allows you to communicate
		// with the user. Note that events should be used sparingly and are subject
		// to throttling; see the issue below for more information.
		// https://github.com/crossplane/crossplane/issues/5802
		response.Warning(rsp, errors.New("something went wrong")).
			TargetCompositeAndClaim()
	} else {
		response.ConditionTrue(rsp, "FunctionSuccess", "Success").
			TargetCompositeAndClaim()
	}

	j, _ := protojson.Marshal(rsp)
	fmt.Println(string(j))

	// Output:
	// {"meta":{"ttl":"60s"},"desired":{"resources":{"new":{"resource":{"apiVersion":"example.org/v1","kind":"CoolResource","metadata":{"labels":{"coolness":"high"}},"spec":{"widgets":9001}}}}},"conditions":[{"type":"FunctionSuccess","status":"STATUS_CONDITION_TRUE","reason":"Success","target":"TARGET_COMPOSITE_AND_CLAIM"}]}
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

// TestMetricsServer_WithCustomRegistryAndCustomPort verifies that metrics server starts on custom port with custom registry as input.
func TestMetricsServer_WithCustomRegistryAndCustomPort(t *testing.T) {
	// Create mock server
	mockServer := &MockFunctionServer{
		rsp: &v1.RunFunctionResponse{
			Meta: &v1.ResponseMeta{Tag: "traffic-test"},
		},
	}

	// Get ports
	grpcPort := getAvailablePort(t)
	metricsPort := getAvailablePort(t)

	// Start server
	serverDone := make(chan error, 1)
	go func() {
		err := Serve(mockServer,
			Listen("tcp", fmt.Sprintf(":%d", grpcPort)),
			Insecure(true),
			WithMetricsServer(fmt.Sprintf(":%d", metricsPort)),
			WithMetricsRegistry(prometheus.NewRegistry()),
		)
		serverDone <- err
	}()

	// Wait for server to start
	time.Sleep(3 * time.Second)

	t.Run("MetricsServerTest On CustomPort With CustomRegistry", func(t *testing.T) {
		conn, err := grpc.NewClient(fmt.Sprintf("localhost:%d", grpcPort),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		client := v1.NewFunctionRunnerServiceClient(conn)

		// Make the request
		req1 := &v1.RunFunctionRequest{
			Meta: &v1.RequestMeta{Tag: "request-test"},
		}

		_, err = client.RunFunction(context.Background(), req1)
		if err != nil {
			t.Errorf("request failed: %v", err)
		}
		// Wait for metrics to be collected
		time.Sleep(2 * time.Second)

		// Verify metrics endpoint has our custom metrics
		metricsURL := fmt.Sprintf("http://localhost:%d/metrics", metricsPort)
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, metricsURL, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to get metrics: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read metrics: %v", err)
		}

		metricsContent := string(body)

		// Verify Prometheus format is working
		if !strings.Contains(metricsContent, "# HELP") {
			t.Error("Expected Prometheus format")
		}

		// Verify gRPC metrics are present
		if !strings.Contains(metricsContent, "grpc_server_started_total") {
			t.Error("Expected grpc_server_started_total metric to be present")
		}
	})
}

// TestMetricsServer_WithDefaultRegistryAndDefaultPort verifies that metrics server starts by default on :8080 with default registry with no input.
func TestMetricsServer_WithDefaultRegistryAndDefaultPort(t *testing.T) {
	// Create mock server
	mockServer := &MockFunctionServer{
		rsp: &v1.RunFunctionResponse{
			Meta: &v1.ResponseMeta{Tag: "default-metrics-test"},
		},
	}

	// Get ports
	grpcPort := getAvailablePort(t)
	// Should use default metrics port 8080
	metricsPort := 8080

	serverDone := make(chan error, 1)
	go func() {
		err := Serve(mockServer,
			Listen("tcp", fmt.Sprintf(":%d", grpcPort)),
			Insecure(true),
		)
		serverDone <- err
	}()

	// Wait for server to start
	time.Sleep(3 * time.Second)

	t.Run("MetricsServerTest On DefaultPort With DefaultRegisrty", func(t *testing.T) {
		// Test gRPC connection
		conn, err := grpc.NewClient(fmt.Sprintf("localhost:%d", grpcPort),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		client := v1.NewFunctionRunnerServiceClient(conn)

		// Make the request
		req := &v1.RunFunctionRequest{
			Meta: &v1.RequestMeta{Tag: "default-metrics-test"},
		}

		_, err = client.RunFunction(context.Background(), req)
		if err != nil {
			t.Errorf("Request failed: %v", err)
		}

		// Wait for metrics to be collected
		time.Sleep(2 * time.Second)

		// Verify metrics endpoint is accessible
		metricsURL := fmt.Sprintf("http://localhost:%d/metrics", metricsPort)
		httpReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, metricsURL, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			t.Fatalf("Failed to get metrics: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read metrics: %v", err)
		}

		metricsContent := string(body)

		// Verify metrics are present
		if !strings.Contains(metricsContent, "# HELP") {
			t.Error("Expected Prometheus format")
		}

		// Verify gRPC metrics are present
		if !strings.Contains(metricsContent, "grpc_server_started_total") {
			t.Error("Expected grpc_server_started_total metric to be present")
		}
	})
}

// Helper function to get an available port.
func getAvailablePort(t *testing.T) int {
	t.Helper()

	listenConfig := &net.ListenConfig{}
	listener, err := listenConfig.Listen(context.Background(), "tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to get available port: %v", err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}
