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

package composed

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/upbound/provider-aws/apis/s3/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"
)

func Example() {
	// Create a new, empty composed resource.
	cd := New()

	// Set our composed resource's type metadata.
	cd.SetAPIVersion("example.org/v1")
	cd.SetKind("CoolComposedResource")

	// Set our composed resource's object metadata.
	cd.SetLabels(map[string]string{"coolness": "high"})

	// Set an arbitrary spec field.
	cd.SetInteger("spec.coolness", 9001)

	// Marshal our composed resource to YAML. We just do this for illustration
	// purposes. Normally you'd add it to the map of desired resources you send
	// in your RunFunctionResponse.
	y, _ := yaml.Marshal(cd)

	fmt.Println(string(y))

	// Output:
	// apiVersion: example.org/v1
	// kind: CoolComposedResource
	// metadata:
	//   labels:
	//     coolness: high
	// spec:
	//   coolness: 9001
}

func ExampleScheme() {
	// Add all v1beta2 types to the scheme so that From can automatically
	// determine their apiVersion and kind.
	v1beta2.AddToScheme(Scheme)
}

func ExampleFrom() {
	// Add all v1beta2 types to the scheme so that From can automatically
	// determine their apiVersion and kind.
	v1beta2.AddToScheme(Scheme)

	// Create a strongly typed runtime.Object, imported from a provider.
	b := &v1beta2.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"coolness": "high",
			},
		},
		Spec: v1beta2.BucketSpec{
			ForProvider: v1beta2.BucketParameters{
				Region: ptr.To[string]("us-east-2"),
			},
		},
	}

	// Create a composed resource from the runtime.Object.
	cd, err := From(b)
	if err != nil {
		panic(err)
	}

	// Marshal our composed resource to YAML. We just do this for illustration
	// purposes. Normally you'd add it to the map of desired resources you send
	// in your RunFunctionResponse.
	y, _ := yaml.Marshal(cd)

	fmt.Println(string(y))

	// Output:
	// apiVersion: s3.aws.upbound.io/v1beta2
	// kind: Bucket
	// metadata:
	//   labels:
	//     coolness: high
	// spec:
	//   forProvider:
	//     region: us-east-2
	// status:
	//   observedGeneration: 0
}

func TestFrom(t *testing.T) {
	v1beta2.AddToScheme(Scheme)

	type args struct {
		o runtime.Object
	}
	type want struct {
		cd  *Unstructured
		err error
	}
	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"WithMetadata": {
			reason: "A resource with metadata should not grow any extra metadata fields during conversion",
			args: args{
				o: &v1beta2.Bucket{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cool-bucket",
					},
					Spec: v1beta2.BucketSpec{
						ForProvider: v1beta2.BucketParameters{
							Region: ptr.To[string]("us-east-2"),
						},
					},
				},
			},
			want: want{
				cd: &Unstructured{Unstructured: unstructured.Unstructured{Object: map[string]any{
					"apiVersion": v1beta2.CRDGroupVersion.String(),
					"kind":       v1beta2.Bucket_Kind,
					"metadata": map[string]any{
						"name": "cool-bucket",
					},
					"spec": map[string]any{
						"forProvider": map[string]any{
							"region": "us-east-2",
						},
					},
					"status": map[string]any{
						"observedGeneration": float64(0),
					},
				}}},
			},
		},
		"WithoutMetadata": {
			reason: "A resource with no metadata should not grow a metadata object during conversion",
			args: args{
				o: &v1beta2.Bucket{
					Spec: v1beta2.BucketSpec{
						ForProvider: v1beta2.BucketParameters{
							Region: ptr.To[string]("us-east-2"),
						},
					},
				},
			},
			want: want{
				cd: &Unstructured{Unstructured: unstructured.Unstructured{Object: map[string]any{
					"apiVersion": v1beta2.CRDGroupVersion.String(),
					"kind":       v1beta2.Bucket_Kind,
					"spec": map[string]any{
						"forProvider": map[string]any{
							"region": "us-east-2",
						},
					},
					"status": map[string]any{
						"observedGeneration": float64(0),
					},
				}}},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cd, err := From(tc.args.o)

			if diff := cmp.Diff(tc.want.cd, cd); diff != "" {
				t.Errorf("\n%s\nFrom(...): -want, +got:\n%s", tc.reason, diff)
			}

			if diff := cmp.Diff(tc.want.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nFrom(...): -want error, +got error:\n%s", tc.reason, diff)
			}
		})
	}
}

func ExampleTo() {
	// Add all v1beta2 types to the scheme so that From can automatically
	// determine their apiVersion and kind.
	v1beta2.AddToScheme(Scheme)

	// Create a unstructured object as we would receive by the function (observed/desired).
	ub := &Unstructured{Unstructured: unstructured.Unstructured{Object: map[string]any{
		"apiVersion": v1beta2.CRDGroupVersion.String(),
		"kind":       v1beta2.Bucket_Kind,
		"metadata": map[string]any{
			"name": "cool-bucket",
		},
		"spec": map[string]any{
			"forProvider": map[string]any{
				"region": "us-east-2",
			},
		},
		"status": map[string]any{
			"observedGeneration": float64(0),
		},
	}}}

	// Create a strongly typed object from the unstructured object.
	sb := &v1beta2.Bucket{}
	err := To(ub, sb)
	if err != nil {
		panic(err)
	}
	// Now you have a strongly typed Bucket object.
	objectLock := true
	sb.Spec.ForProvider.ObjectLockEnabled = &objectLock
}

// Test the To function
func TestTo(t *testing.T) {
	v1beta2.AddToScheme(Scheme)
	type args struct {
		un  *Unstructured
		obj interface{}
	}
	type want struct {
		obj interface{}
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"SuccessfulConversion": {
			reason: "A valid unstructured object should convert to a structured object without errors",
			args: args{
				un: &Unstructured{Unstructured: unstructured.Unstructured{Object: map[string]any{
					"apiVersion": v1beta2.CRDGroupVersion.String(),
					"kind":       v1beta2.Bucket_Kind,
					"metadata": map[string]any{
						"name": "cool-bucket",
					},
					"spec": map[string]any{
						"forProvider": map[string]any{
							"region": "us-east-2",
						},
					},
					"status": map[string]any{
						"observedGeneration": float64(0),
					},
				}}},
				obj: &v1beta2.Bucket{},
			},
			want: want{
				obj: &v1beta2.Bucket{
					TypeMeta: metav1.TypeMeta{
						Kind:       v1beta2.Bucket_Kind,
						APIVersion: v1beta2.CRDGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "cool-bucket",
					},
					Spec: v1beta2.BucketSpec{
						ForProvider: v1beta2.BucketParameters{
							Region: ptr.To[string]("us-east-2"),
						},
					},
				},
				err: nil,
			},
		},
		"InvalidGVK": {
			reason: "An unstructured object with mismatched GVK should result in an error",
			args: args{
				un: &Unstructured{Unstructured: unstructured.Unstructured{Object: map[string]any{
					"apiVersion": "test.example.io",
					"kind":       "Unknown",
					"metadata": map[string]any{
						"name": "cool-bucket",
					},
					"spec": map[string]any{
						"forProvider": map[string]any{
							"region": "us-east-2",
						},
					},
					"status": map[string]any{
						"observedGeneration": float64(0),
					},
				}}},
				obj: &v1beta2.Bucket{},
			},
			want: want{
				obj: &v1beta2.Bucket{},
				err: errors.New("GVK /test.example.io, Kind=Unknown is not known by the scheme for the provided object type"),
			},
		},
		"NoRuntimeObject": {
			reason: "Should only convert to a object if the object is a runtime.Object",
			args: args{
				un: &Unstructured{Unstructured: unstructured.Unstructured{Object: map[string]any{
					"apiVersion": v1beta1.CRDGroupVersion.String(),
					"kind":       v1beta1.Bucket_Kind,
					"metadata": map[string]any{
						"name": "cool-bucket",
					},
				}}},
				obj: "not-a-runtime-object",
			},
			want: want{
				obj: string("not-a-runtime-object"),
				err: errors.New("object is not a compatible runtime.Object"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := To(tc.args.un, tc.args.obj)

			// Compare the resulting object with the expected one
			if diff := cmp.Diff(tc.want.obj, tc.args.obj); diff != "" {
				t.Errorf("\n%s\nTo(...): -want, +got:\n%s", tc.reason, diff)
			}
			// Compare the error with the expected error
			if diff := cmp.Diff(tc.want.err, err, EquateErrors()); diff != "" {
				t.Errorf("\n%s\nTo(...): -want error, +got error:\n%s", tc.reason, diff)
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
