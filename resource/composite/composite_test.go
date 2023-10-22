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

package composite

import (
	"errors"
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/util/yaml"
)

var manifest = []byte(`
apiVersion: example.org/v1
kind: CoolCompositeResource
metadata:
  name: my-cool-xr
spec:
  intfield: 9001
  floatfield: 9.001
  floatexp: 10e3
  notanint: foo
`)

func TestGetInteger(t *testing.T) {
	errNotFound := func(path string) error {
		p := &fieldpath.Paved{}
		_, err := p.GetValue(path)
		return err
	}
	// Create a new, empty XR.
	xr := New()

	// Unmarshal our manifest into the XR. This is just for illustration
	// purposes - the SDK functions like GetObservedCompositeResource do this
	// for you.
	_ = yaml.Unmarshal(manifest, xr)

	type args struct {
		path string
	}
	type want struct {
		i   int64
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"Integer": {
			reason: "Correctly return an integer",
			args: args{
				path: "spec.intfield",
			},
			want: want{
				i:   9001,
				err: nil,
			},
		},
		"Float": {
			reason: "Correctly truncates a float",
			args: args{
				path: "spec.floatfield",
			},
			want: want{
				i:   9,
				err: nil,
			},
		},
		"FloatExp": {
			reason: "Correctly truncates a float with an exponent",
			args: args{
				path: "spec.floatexp",
			},
			want: want{
				i:   10000,
				err: nil,
			},
		},
		"ParseString": {
			reason: "Error when not a number",
			args: args{
				path: "spec.notanint",
			},
			want: want{
				i:   0,
				err: errors.New("spec.notanint: not a (int64) number"),
			},
		},
		"MissingField": {
			reason: "Return 0 and error on a missing field",
			args: args{
				path: "badfield",
			},
			want: want{
				i:   0,
				err: errNotFound("badfield"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			i, err := xr.GetInteger(tc.args.path)

			if diff := cmp.Diff(tc.want.i, i); diff != "" {
				t.Errorf("%s\nGetInteger(...): -want i, +got i:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.err, err, EquateErrors()); diff != "" {
				t.Errorf("%s\nGetInteger(...): -want err, +got err:\n%s", tc.reason, diff)
			}
		})
	}

}

// EquateErrors returns true if the supplied errors are of the same type and
// produce identical strings. This mirrors the error comparison behaviour of
// https://github.com/go-test/deep,
//
// This differs from cmpopts.EquateErrors, which does not test for error strings
// and instead returns whether one error 'is' (in the errors.Is sense) the
// other.
func EquateErrors() cmp.Option {
	return cmp.Comparer(func(a, b error) bool {
		if a == nil || b == nil {
			return a == nil && b == nil
		}
		return a.Error() == b.Error()
	})
}
