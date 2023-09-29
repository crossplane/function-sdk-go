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
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composed"
	"github.com/crossplane/function-sdk-go/response"
)

var req = &v1beta1.RunFunctionRequest{
	Observed: &v1beta1.State{
		Composite: &v1beta1.Resource{
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
	// db, _ := composed.From(&v1beta1.Instance{})
	// desired["database"] = &resource.DesiredComposed{Resource: db}

	// Add a label to our new desired resource, and any other.
	for _, r := range desired {
		r.Resource.SetLabels(map[string]string{"coolness": "high"})
	}

	// Set our updated desired composed resource in the response we'll return.
	_ = response.SetDesiredComposedResources(rsp, desired)

	j, _ := protojson.Marshal(rsp)
	fmt.Println(string(j))

	// Output:
	// {"meta":{"ttl":"60s"}, "desired":{"resources":{"new":{"resource":{"apiVersion":"example.org/v1", "kind":"CoolResource", "metadata":{"labels":{"coolness":"high"}}, "spec":{"widgets":9001}}}}}}
}
