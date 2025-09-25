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

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	v1 "github.com/crossplane/function-sdk-go/proto/v1"
)

// TestMetricsWithRealTraffic tests metrics collection with real traffic patterns.
func TestMetricsWithRealTraffic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Use default registry - the gRPC interceptor will register its metrics there
	// We don't need a custom registry for this test

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
		)
		serverDone <- err
	}()

	// Wait for server to start
	time.Sleep(3 * time.Second)

	// Test with two requests
	t.Run("TwoRequestsTest", func(t *testing.T) {
		conn, err := grpc.NewClient(fmt.Sprintf("localhost:%d", grpcPort),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		client := v1.NewFunctionRunnerServiceClient(conn)

		// Make first request
		req1 := &v1.RunFunctionRequest{
			Meta: &v1.RequestMeta{Tag: "first-request-test"},
		}

		_, err = client.RunFunction(context.Background(), req1)
		if err != nil {
			t.Errorf("First request failed: %v", err)
		}

		// Make second request
		req2 := &v1.RunFunctionRequest{
			Meta: &v1.RequestMeta{Tag: "second-request-test"},
		}

		_, err = client.RunFunction(context.Background(), req2)
		if err != nil {
			t.Errorf("Second request failed: %v", err)
		}

		// The gRPC interceptor should automatically record metrics for both requests
		// We don't need to manually increment counters - the interceptor handles this

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

		// Assert gRPC interceptor metrics are present and show 2 requests
		if !strings.Contains(metricsContent, "grpc_server_started_total") {
			t.Error("Expected grpc_server_started_total metric to be present")
		}

		if !strings.Contains(metricsContent, "grpc_server_handled_total") {
			t.Error("Expected grpc_server_handled_total metric to be present")
		}

		// Assert that exactly 2 requests were started
		if !strings.Contains(metricsContent, `grpc_server_started_total{grpc_method="RunFunction",grpc_service="apiextensions.fn.proto.v1.FunctionRunnerService",grpc_type="unary"} 2`) {
			t.Error("Expected grpc_server_started_total to show 2 requests for v1.FunctionRunnerService")
		}

		// Assert that exactly 2 requests were handled successfully
		if !strings.Contains(metricsContent, `grpc_server_handled_total{grpc_code="OK",grpc_method="RunFunction",grpc_service="apiextensions.fn.proto.v1.FunctionRunnerService",grpc_type="unary"} 2`) {
			t.Error("Expected grpc_server_handled_total to show 2 successful requests for v1.FunctionRunnerService")
		}

		// Assert that messages were received and sent
		if !strings.Contains(metricsContent, `grpc_server_msg_received_total{grpc_method="RunFunction",grpc_service="apiextensions.fn.proto.v1.FunctionRunnerService",grpc_type="unary"} 2`) {
			t.Error("Expected grpc_server_msg_received_total to show 2 messages received")
		}

		if !strings.Contains(metricsContent, `grpc_server_msg_sent_total{grpc_method="RunFunction",grpc_service="apiextensions.fn.proto.v1.FunctionRunnerService",grpc_type="unary"} 2`) {
			t.Error("Expected grpc_server_msg_sent_total to show 2 messages sent")
		}

		// Assert Prometheus format is correct
		if !strings.Contains(metricsContent, "# HELP") {
			t.Error("Expected Prometheus format with HELP comments")
		}

		// Assert Go runtime metrics are present (verifies metrics server is working)
		if !strings.Contains(metricsContent, "go_goroutines") {
			t.Error("Expected Go runtime metrics to be present")
		}
	})
}

// Helper function to get an available port.
func getAvailablePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to get available port: %v", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	return port
}
