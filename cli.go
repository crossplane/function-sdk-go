/*
Copyright 2026 The Crossplane Authors.

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
	"github.com/alecthomas/kong"

	"github.com/crossplane/function-sdk-go/logging"
)

// CLI provides standard flags and environment variables for Composition
// Functions. It is designed to be used with [github.com/alecthomas/kong].
//
// Without custom flags, use CLI directly with [Parse]:
//
//	type CLI struct {
//	    function.CLI `kong:"embed"`
//	    // MyFlag string `default:"foo" env:"MY_FLAG" help:"My custom flag."`
//	}
//
//	func (c *CLI) Run() error {
//	    log, err := c.Logger()
//	    if err != nil {
//	        return err
//	    }
//	    return function.Serve(&Function{log: log}, c.StandardOptions()...)
//	    // or with custom flags:
//	    // return function.Serve(&Function{log: log, myFlag: c.MyFlag}, c.StandardOptions()...)
//	}
//
//	func main() {
//	    function.Parse(&CLI{}, "My function.")
//	}
type CLI struct {
	Address            string `default:":9443"            env:"ADDRESS"                                                                                                       help:"Address at which to listen for gRPC connections."`
	Debug              bool   `env:"DEBUG"                help:"Emit debug logs in addition to info logs."                                                                    short:"d"`
	Insecure           bool   `env:"INSECURE"             help:"Run without mTLS credentials. If you supply this flag --tls-server-certs-dir will be ignored."`
	MaxRecvMessageSize int    `default:"4"                env:"MAX_RECV_MESSAGE_SIZE"                                                                                         help:"Maximum size of received messages in MB."`
	Network            string `default:"tcp"              env:"NETWORK"                                                                                                       help:"Network on which to listen for gRPC connections."`
	TLSCertsDir        string `env:"TLS_SERVER_CERTS_DIR" help:"Directory containing server certs (tls.key, tls.crt) and the CA used to verify client certificates (ca.crt)."`
}

// StandardOptions returns the ServeOptions derived from standard CLI flags.
func (c *CLI) StandardOptions() []ServeOption {
	return []ServeOption{
		Listen(c.Network, c.Address),
		MTLSCertificates(c.TLSCertsDir),
		Insecure(c.Insecure),
		MaxRecvMessageSize(c.MaxRecvMessageSize * 1024 * 1024),
	}
}

// Logger returns a new logger configured from CLI flags.
func (c *CLI) Logger() (logging.Logger, error) {
	return NewLogger(c.Debug)
}

// Parse parses CLI flags using kong and runs the command. The cli argument must
// have a Run() error method. An optional description is used as CLI help text.
func Parse(cli any, description ...string) {
	options := []kong.Option{}
	if len(description) > 0 {
		options = append(options, kong.Description(description[0]))
	}
	ctx := kong.Parse(cli, options...)
	ctx.FatalIfErrorf(ctx.Run())
}
