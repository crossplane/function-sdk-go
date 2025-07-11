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

// Package function is an SDK for building Composition Functions.
package function

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	ginsecure "google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/proto"

	"github.com/crossplane/function-sdk-go/logging"
	v1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/proto/v1beta1"
)

// Default ServeOptions.
const (
	DefaultNetwork        = "tcp"
	DefaultAddress        = ":9443"
	DefaultMaxRecvMsgSize = 1024 * 1024 * 4
)

// ServeOptions configure how a Function is served.
type ServeOptions struct {
	Network        string
	Address        string
	MaxRecvMsgSize int
	Credentials    credentials.TransportCredentials
}

// A ServeOption configures how a Function is served.
type ServeOption func(o *ServeOptions) error

// Listen configures the network, address, and maximum message size on which the
// Function will listen for RunFunctionRequests.
func Listen(network, address string) ServeOption {
	return func(o *ServeOptions) error {
		o.Network = network
		o.Address = address
		return nil
	}
}

// MTLSCertificates specifies a directory from which to load mTLS certificates.
// The directory must contain the server certificate (tls.key and tls.crt), as
// well as a CA certificate (ca.crt) that will be used to authenticate clients.
func MTLSCertificates(dir string) ServeOption {
	return func(o *ServeOptions) error {
		if dir == "" {
			// We want to support passing both MTLSCertificates and
			// Insecure as they were supplied as flags. So we don't
			// want this to fail because no dir was supplied.
			// If no TLS dir is supplied and insecure is false we'll
			// return an error due to having no credentials specified.
			return nil
		}
		crt, err := tls.LoadX509KeyPair(
			filepath.Clean(filepath.Join(dir, "tls.crt")),
			filepath.Clean(filepath.Join(dir, "tls.key")),
		)
		if err != nil {
			return errors.Wrap(err, "cannot load X509 keypair")
		}

		ca, err := os.ReadFile(filepath.Clean(filepath.Join(dir, "ca.crt")))
		if err != nil {
			return errors.Wrap(err, "cannot read CA certificate")
		}

		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(ca) {
			return errors.New("invalid CA certificate")
		}

		o.Credentials = credentials.NewTLS(&tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{crt},
			ClientCAs:    pool,
			ClientAuth:   tls.RequireAndVerifyClientCert,
		})

		return nil
	}
}

// Insecure specifies whether this Function should be served insecurely - i.e.
// without mTLS authentication. This is only useful for testing and development.
// Crossplane will always send requests using mTLS.
func Insecure(insecure bool) ServeOption {
	return func(o *ServeOptions) error {
		if insecure {
			o.Credentials = ginsecure.NewCredentials()
		}
		return nil
	}
}

// MaxRecvMessageSize returns a ServeOption to set the max message size in bytes the server can receive.
// If this is not set, gRPC uses the default limit.
func MaxRecvMessageSize(sz int) ServeOption {
	return func(o *ServeOptions) error {
		o.MaxRecvMsgSize = sz
		return nil
	}
}

// Serve the supplied Function by creating a gRPC server and listening for
// RunFunctionRequests. Blocks until the server returns an error.
func Serve(fn v1.FunctionRunnerServiceServer, o ...ServeOption) error {
	so := &ServeOptions{
		Network:        DefaultNetwork,
		Address:        DefaultAddress,
		MaxRecvMsgSize: DefaultMaxRecvMsgSize,
	}

	for _, fn := range o {
		if err := fn(so); err != nil {
			return errors.Wrap(err, "cannot apply ServeOption")
		}
	}

	if so.Credentials == nil {
		return errors.New("no credentials provided - did you specify the Insecure or MTLSCertificates options?")
	}

	lis, err := net.Listen(so.Network, so.Address)
	if err != nil {
		return errors.Wrapf(err, "cannot listen for %s connections at address %q", so.Network, so.Address)
	}

	srv := grpc.NewServer(grpc.MaxRecvMsgSize(so.MaxRecvMsgSize), grpc.Creds(so.Credentials))
	reflection.Register(srv)
	v1.RegisterFunctionRunnerServiceServer(srv, fn)
	v1beta1.RegisterFunctionRunnerServiceServer(srv, ServeBeta(fn))
	return errors.Wrap(srv.Serve(lis), "cannot serve mTLS gRPC connections")
}

// NewLogger returns a new logger.
func NewLogger(debug bool) (logging.Logger, error) {
	return logging.NewLogger(debug)
}

// A BetaServer is a v1beta1 FunctionRunnerServiceServer that wraps an identical
// v1 FunctionRunnerServiceServer. This requires the v1 and v1beta1 protos to be
// identical.
//
// Functions were promoted from v1beta1 to v1 in Crossplane v1.17. Crossplane
// v1.16 and earlier only sends v1beta1 RunFunctionRequests. Functions should
// use the BetaServer for backward compatibility, to support Crossplane v1.16
// and earlier.
type BetaServer struct {
	v1beta1.UnimplementedFunctionRunnerServiceServer

	wrapped v1.FunctionRunnerServiceServer
}

// ServeBeta returns a v1beta1.FunctionRunnerServiceServer that wraps the
// suppled v1.FunctionRunnerServiceServer.
func ServeBeta(s v1.FunctionRunnerServiceServer) *BetaServer {
	return &BetaServer{wrapped: s}
}

// RunFunction calls the RunFunction method of the wrapped
// v1.FunctionRunnerServiceServer. It converts from v1beta1 to v1 and back by
// round-tripping through protobuf marshaling.
func (s *BetaServer) RunFunction(ctx context.Context, req *v1beta1.RunFunctionRequest) (*v1beta1.RunFunctionResponse, error) {
	gareq := &v1.RunFunctionRequest{}

	b, err := proto.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "cannot marshal v1beta1 RunFunctionRequest to protobuf bytes")
	}

	if err := proto.Unmarshal(b, gareq); err != nil {
		return nil, errors.Wrap(err, "cannot unmarshal v1 RunFunctionRequest from v1beta1 protobuf bytes")
	}

	garsp, err := s.wrapped.RunFunction(ctx, gareq)
	if err != nil {
		// This error is intentionally not wrapped. This middleware is just
		// calling an underlying RunFunction.
		return nil, err
	}

	b, err = proto.Marshal(garsp)
	if err != nil {
		return nil, errors.Wrap(err, "cannot marshal v1beta1 RunFunctionResponse to protobuf bytes")
	}

	rsp := &v1beta1.RunFunctionResponse{}
	err = proto.Unmarshal(b, rsp)
	return rsp, errors.Wrap(err, "cannot unmarshal v1 RunFunctionResponse from v1beta1 protobuf bytes")
}
