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
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/upbound/provider-aws/apis/s3/v1beta1"
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
	// Add all v1beta1 types to the scheme so that From can automatically
	// determine their apiVersion and kind.
	v1beta1.AddToScheme(Scheme)
}

func ExampleFrom() {
	// Add all v1beta1 types to the scheme so that From can automatically
	// determine their apiVersion and kind.
	v1beta1.AddToScheme(Scheme)

	// Create a strongly typed runtime.Object, imported from a provider.
	b := &v1beta1.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"coolness": "high",
			},
		},
		Spec: v1beta1.BucketSpec{
			ForProvider: v1beta1.BucketParameters{
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
	// apiVersion: s3.aws.upbound.io/v1beta1
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
	v1beta1.AddToScheme(Scheme)

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
				o: &v1beta1.Bucket{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cool-bucket",
					},
					Spec: v1beta1.BucketSpec{
						ForProvider: v1beta1.BucketParameters{
							Region: ptr.To[string]("us-east-2"),
						},
					},
				},
			},
			want: want{
				cd: &Unstructured{Unstructured: unstructured.Unstructured{Object: map[string]any{
					"apiVersion": v1beta1.CRDGroupVersion.String(),
					"kind":       v1beta1.Bucket_Kind,
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
				o: &v1beta1.Bucket{
					Spec: v1beta1.BucketSpec{
						ForProvider: v1beta1.BucketParameters{
							Region: ptr.To[string]("us-east-2"),
						},
					},
				},
			},
			want: want{
				cd: &Unstructured{Unstructured: unstructured.Unstructured{Object: map[string]any{
					"apiVersion": v1beta1.CRDGroupVersion.String(),
					"kind":       v1beta1.Bucket_Kind,
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
