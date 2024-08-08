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

	"github.com/crossplane/function-sdk-go/logging"
	"github.com/crossplane/function-sdk-go/proto/v1beta1"
)

// Default ServeOptions.
const (
	DefaultNetwork = "tcp"
	DefaultAddress = ":9443"
)

// ServeOptions configure how a Function is served.
type ServeOptions struct {
	Network     string
	Address     string
	Credentials credentials.TransportCredentials
}

// A ServeOption configures how a Function is served.
type ServeOption func(o *ServeOptions) error

// Listen configures the network and address on which the Function will
// listen for RunFunctionRequests.
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

// Serve the supplied Function by creating a gRPC server and listening for
// RunFunctionRequests. Blocks until the server returns an error.
func Serve(fn v1beta1.FunctionRunnerServiceServer, o ...ServeOption) error {
	so := &ServeOptions{
		Network: DefaultNetwork,
		Address: DefaultAddress,
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

	srv := grpc.NewServer(grpc.Creds(so.Credentials))
	reflection.Register(srv)
	v1beta1.RegisterFunctionRunnerServiceServer(srv, fn)
	return errors.Wrap(srv.Serve(lis), "cannot serve mTLS gRPC connections")
}

// NewLogger returns a new logger.
func NewLogger(debug, timeEncodeISO8601 bool, addCallerSkip ...int) (logging.Logger, error) {
	return logging.NewLogger(debug, timeEncodeISO8601, addCallerSkip...)
}
