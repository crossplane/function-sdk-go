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

package resource

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/crossplane/function-sdk-go/resource/composed"
)

func TestAsObjectUnstructuredNumbers(t *testing.T) {
	type want struct {
		generation int64
		object     map[string]any
	}

	cases := map[string]struct {
		reason string
		s      map[string]any
		want   want
	}{
		"WholeNumbersBecomeInt64": {
			reason: "Whole numbers must be stored as int64 so that apimachinery's accessors, such as GetGeneration, can read them.",
			s: map[string]any{
				"apiVersion": "example.org/v1",
				"kind":       "Example",
				"metadata":   map[string]any{"name": "a", "generation": 8},
				"spec":       map[string]any{"replicas": 3, "weights": []any{1, 2.5}},
			},
			want: want{
				generation: 8,
				object: map[string]any{
					"apiVersion": "example.org/v1",
					"kind":       "Example",
					"metadata":   map[string]any{"name": "a", "generation": int64(8)},
					"spec":       map[string]any{"replicas": int64(3), "weights": []any{int64(1), 2.5}},
				},
			},
		},
		"FractionsStayFloat64": {
			reason: "A number with a fractional part has no int64 form and must be left alone.",
			s: map[string]any{
				"apiVersion": "example.org/v1",
				"kind":       "Example",
				"metadata":   map[string]any{"name": "a"},
				"spec":       map[string]any{"ratio": 0.5},
			},
			want: want{
				object: map[string]any{
					"apiVersion": "example.org/v1",
					"kind":       "Example",
					"metadata":   map[string]any{"name": "a"},
					"spec":       map[string]any{"ratio": 0.5},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			s, err := structpb.NewStruct(tc.s)
			if err != nil {
				t.Fatalf("structpb.NewStruct(...): %v", err)
			}

			o := composed.New()
			if err := AsObject(s, o); err != nil {
				t.Fatalf("AsObject(...): %v", err)
			}

			if diff := cmp.Diff(tc.want.generation, o.GetGeneration()); diff != "" {
				t.Errorf("\n%s\nAsObject(...): -want generation, +got generation:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.object, o.UnstructuredContent()); diff != "" {
				t.Errorf("\n%s\nAsObject(...): -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}
