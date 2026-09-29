# function-sdk-go
[![CI](https://github.com/crossplane/function-sdk-go/actions/workflows/ci.yml/badge.svg)](https://github.com/crossplane/function-sdk-go/actions/workflows/ci.yml) ![GitHub release (latest SemVer)](https://img.shields.io/github/release/crossplane/function-sdk-go) [![Go Reference](https://pkg.go.dev/badge/github.com/crossplane/function-sdk-go.svg)](https://pkg.go.dev/github.com/crossplane/function-sdk-go)

The [Go][go] SDK for writing [composition functions][functions].

This SDK is currently a beta. We try to avoid breaking changes, but it will not
have a stable API until it reaches v1.0.0. It follows the same [contributing
guidelines] as Crossplane.

To learn how to use this SDK:

* [Follow the guide to writing a composition function in Go][function guide]
* [Learn about how composition functions work][functions]
* [Read the package documentation][package docs]

If you just want to jump in and get started, consider using the
[function-template-go] template repository.

## Standard CLI options

Embed `function.CLI` to get the standard flags and environment variables
(`--address`, `--debug`, `--insecure`, `--tls-certs-dir`, `--ttl`, and more).
Add your own flags alongside it:

```go
type CLI struct {
	function.CLI `kong:"embed"`

	MyFlag string `default:"foo" env:"MY_FLAG" help:"My custom flag."`
}

func (c *CLI) Run() error {
	log, err := c.Logger()
	if err != nil {
		return err
	}
	// Use c.TTL with response.To(req, c.TTL) in RunFunction.
	return function.Serve(&Function{log: log, ttl: c.TTL, myFlag: c.MyFlag}, c.StandardOptions()...)
}

func main() {
	function.Parse(&CLI{}, "My function.")
}
```

See the [package docs] for the full list of options.

[go]: https://go.dev
[functions]: https://docs.crossplane.io/latest/concepts/composition-functions
[contributing guidelines]: https://github.com/crossplane/crossplane/tree/master/contributing
[function guide]: https://docs.crossplane.io/knowledge-base/guides/write-a-composition-function-in-go
[package docs]: https://pkg.go.dev/github.com/crossplane/function-sdk-go
[function-template-go]: https://github.com/crossplane/function-template-go
