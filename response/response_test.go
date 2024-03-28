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

package response_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/response"
)

func TestResult(t *testing.T) {
	type testFn func(*v1beta1.RunFunctionResponse)
	type args struct {
		fns []testFn
	}
	type want struct {
		results []*v1beta1.Result
	}
	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"CreateBasicRecords": {
			reason: "Correctly adds results to the response.",
			args: args{
				fns: []testFn{
					func(rsp *v1beta1.RunFunctionResponse) {
						response.Normal(rsp, "this is a test normal result")
					},
					func(rsp *v1beta1.RunFunctionResponse) {
						response.Warning(rsp, errors.New("this is a test warning result"))
					},
					func(rsp *v1beta1.RunFunctionResponse) {
						response.Fatal(rsp, errors.New("this is a test fatal result"))
					},
				},
			},
			want: want{
				results: []*v1beta1.Result{
					{
						Severity: v1beta1.Severity_SEVERITY_NORMAL,
						Message:  "this is a test normal result",
						Target:   v1beta1.Target_TARGET_COMPOSITE,
					},
					{
						Severity: v1beta1.Severity_SEVERITY_WARNING,
						Message:  "this is a test warning result",
						Target:   v1beta1.Target_TARGET_COMPOSITE,
					},
					{
						Severity: v1beta1.Severity_SEVERITY_FATAL,
						Message:  "this is a test fatal result",
						Target:   v1beta1.Target_TARGET_COMPOSITE,
					},
				},
			},
		},
		"SetTargets": {
			reason: "Correctly sets targets on result and adds it to the response.",
			args: args{
				fns: []testFn{
					func(rsp *v1beta1.RunFunctionResponse) {
						response.Warning(rsp, errors.New("this is a test warning result targeting the composite")).TargetComposite()
					},
					func(rsp *v1beta1.RunFunctionResponse) {
						response.Fatal(rsp, errors.New("this is a test fatal result targeting both")).TargetCompositeAndClaim()
					},
				},
			},
			want: want{
				results: []*v1beta1.Result{
					{
						Severity: v1beta1.Severity_SEVERITY_WARNING,
						Message:  "this is a test warning result targeting the composite",
						Target:   v1beta1.Target_TARGET_COMPOSITE,
					},
					{
						Severity: v1beta1.Severity_SEVERITY_FATAL,
						Message:  "this is a test fatal result targeting both",
						Target:   v1beta1.Target_TARGET_COMPOSITE_AND_CLAIM,
					},
				},
			},
		},
		"CreatesConditions": {
			reason: "Correctly sets targets on result and adds it to the response.",
			args: args{
				fns: []testFn{
					func(rsp *v1beta1.RunFunctionResponse) {
						response.Normal(rsp, "some-message").ConditionTrue("DatabaseReady", "Available")
					},
					func(rsp *v1beta1.RunFunctionResponse) {
						response.Warning(rsp, errors.New("some-error")).ConditionUnknown("DatabaseReady", "Stale")
					},
					func(rsp *v1beta1.RunFunctionResponse) {
						response.Fatal(rsp, errors.New("some-error")).ConditionFalse("DatabaseReady", "Unauthorized")
					},
				},
			},
			want: want{
				results: []*v1beta1.Result{
					{
						Severity: v1beta1.Severity_SEVERITY_NORMAL,
						Message:  "some-message",
						Target:   v1beta1.Target_TARGET_COMPOSITE,
						Condition: &v1beta1.Condition{
							Type:   "DatabaseReady",
							Status: v1beta1.Status_STATUS_TRUE,
							Reason: "Available",
						},
					},
					{
						Severity: v1beta1.Severity_SEVERITY_WARNING,
						Message:  "some-error",
						Target:   v1beta1.Target_TARGET_COMPOSITE,
						Condition: &v1beta1.Condition{
							Type:   "DatabaseReady",
							Status: v1beta1.Status_STATUS_UNKNOWN,
							Reason: "Stale",
						},
					},
					{
						Severity: v1beta1.Severity_SEVERITY_FATAL,
						Message:  "some-error",
						Target:   v1beta1.Target_TARGET_COMPOSITE,
						Condition: &v1beta1.Condition{
							Type:   "DatabaseReady",
							Status: v1beta1.Status_STATUS_FALSE,
							Reason: "Unauthorized",
						},
					},
				},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rsp := &v1beta1.RunFunctionResponse{}
			for _, f := range tc.args.fns {
				f(rsp)
			}

			if diff := cmp.Diff(tc.want.results, rsp.GetResults(), protocmp.Transform()); diff != "" {
				t.Errorf("\n%s\nFrom(...): -want, +got:\n%s", tc.reason, diff)
			}

		})
	}
}
