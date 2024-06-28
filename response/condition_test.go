/*
Copyright 2024 The Crossplane Authors.

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
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	"k8s.io/utils/ptr"

	"github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/response"
)

// Condition types.
const (
	typeDatabaseReady = "DatabaseReady"
)

// Condition reasons.
const (
	reasonAvailable    = "ReasonAvailable"
	reasonCreating     = "ReasonCreating"
	reasonPriorFailure = "ReasonPriorFailure"
	reasonUnauthorized = "ReasonUnauthorized"
)

func TestCondition(t *testing.T) {
	type testFn func(*v1beta1.RunFunctionResponse)
	type args struct {
		fns []testFn
	}
	type want struct {
		conditions []*v1beta1.Condition
	}
	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"CreateBasicRecords": {
			reason: "Correctly adds conditions to the response.",
			args: args{
				fns: []testFn{
					func(rsp *v1beta1.RunFunctionResponse) {
						response.ConditionTrue(rsp, typeDatabaseReady, reasonAvailable)
					},
					func(rsp *v1beta1.RunFunctionResponse) {
						response.ConditionFalse(rsp, typeDatabaseReady, reasonCreating)
					},
					func(rsp *v1beta1.RunFunctionResponse) {
						response.ConditionUnknown(rsp, typeDatabaseReady, reasonPriorFailure)
					},
				},
			},
			want: want{
				conditions: []*v1beta1.Condition{
					{
						Type:   typeDatabaseReady,
						Status: v1beta1.Status_STATUS_CONDITION_TRUE,
						Reason: reasonAvailable,
						Target: v1beta1.Target_TARGET_COMPOSITE.Enum(),
					},
					{
						Type:   typeDatabaseReady,
						Status: v1beta1.Status_STATUS_CONDITION_FALSE,
						Reason: reasonCreating,
						Target: v1beta1.Target_TARGET_COMPOSITE.Enum(),
					},
					{
						Type:   typeDatabaseReady,
						Status: v1beta1.Status_STATUS_CONDITION_UNKNOWN,
						Reason: reasonPriorFailure,
						Target: v1beta1.Target_TARGET_COMPOSITE.Enum(),
					},
				},
			},
		},
		"SetTargets": {
			reason: "Correctly sets targets on condition and adds it to the response.",
			args: args{
				fns: []testFn{
					func(rsp *v1beta1.RunFunctionResponse) {
						response.ConditionTrue(rsp, typeDatabaseReady, reasonAvailable).TargetComposite()
					},
					func(rsp *v1beta1.RunFunctionResponse) {
						response.ConditionTrue(rsp, typeDatabaseReady, reasonAvailable).TargetCompositeAndClaim()
					},
				},
			},
			want: want{
				conditions: []*v1beta1.Condition{
					{
						Type:   typeDatabaseReady,
						Status: v1beta1.Status_STATUS_CONDITION_TRUE,
						Reason: reasonAvailable,
						Target: v1beta1.Target_TARGET_COMPOSITE.Enum(),
					},
					{
						Type:   typeDatabaseReady,
						Status: v1beta1.Status_STATUS_CONDITION_TRUE,
						Reason: reasonAvailable,
						Target: v1beta1.Target_TARGET_COMPOSITE_AND_CLAIM.Enum(),
					},
				},
			},
		},
		"SetMessage": {
			reason: "Correctly sets message on condition and adds it to the response.",
			args: args{
				fns: []testFn{
					func(rsp *v1beta1.RunFunctionResponse) {
						response.ConditionTrue(rsp, typeDatabaseReady, reasonAvailable).WithMessage("a test message")
					},
				},
			},
			want: want{
				conditions: []*v1beta1.Condition{
					{
						Type:    typeDatabaseReady,
						Status:  v1beta1.Status_STATUS_CONDITION_TRUE,
						Reason:  reasonAvailable,
						Target:  v1beta1.Target_TARGET_COMPOSITE.Enum(),
						Message: ptr.To("a test message"),
					},
				},
			},
		},
		"ChainOptions": {
			reason: "Can chain condition options together.",
			args: args{
				fns: []testFn{
					func(rsp *v1beta1.RunFunctionResponse) {
						response.ConditionTrue(rsp, typeDatabaseReady, reasonAvailable).
							WithMessage("a test message").
							TargetCompositeAndClaim()
					},
					func(rsp *v1beta1.RunFunctionResponse) {
						response.ConditionTrue(rsp, typeDatabaseReady, reasonAvailable).
							TargetCompositeAndClaim().
							WithMessage("a test message")
					},
				},
			},
			want: want{
				conditions: []*v1beta1.Condition{
					{
						Type:    typeDatabaseReady,
						Status:  v1beta1.Status_STATUS_CONDITION_TRUE,
						Reason:  reasonAvailable,
						Target:  v1beta1.Target_TARGET_COMPOSITE_AND_CLAIM.Enum(),
						Message: ptr.To("a test message"),
					},
					{
						Type:    typeDatabaseReady,
						Status:  v1beta1.Status_STATUS_CONDITION_TRUE,
						Reason:  reasonAvailable,
						Target:  v1beta1.Target_TARGET_COMPOSITE_AND_CLAIM.Enum(),
						Message: ptr.To("a test message"),
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

			if diff := cmp.Diff(tc.want.conditions, rsp.GetConditions(), protocmp.Transform()); diff != "" {
				t.Errorf("\n%s\nFrom(...): -want, +got:\n%s", tc.reason, diff)
			}

		})
	}
}
