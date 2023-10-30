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

	"github.com/upbound/provider-aws/apis/s3/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func ExampleFrom() {
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
}
