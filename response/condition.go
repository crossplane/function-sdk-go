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

package response

import (
	v1 "github.com/crossplane/function-sdk-go/proto/v1"
)

// ConditionOption allows further customization of the condition.
type ConditionOption struct {
	condition *v1.Condition
}

// ConditionTrue will create a condition with the status of true and add the
// condition to the supplied RunFunctionResponse.
func ConditionTrue(rsp *v1.RunFunctionResponse, typ, reason string) *ConditionOption {
	return newCondition(rsp, typ, reason, v1.Status_STATUS_CONDITION_TRUE)
}

// ConditionFalse will create a condition with the status of false and add the
// condition to the supplied RunFunctionResponse.
func ConditionFalse(rsp *v1.RunFunctionResponse, typ, reason string) *ConditionOption {
	return newCondition(rsp, typ, reason, v1.Status_STATUS_CONDITION_FALSE)
}

// ConditionUnknown will create a condition with the status of unknown and add
// the condition to the supplied RunFunctionResponse.
func ConditionUnknown(rsp *v1.RunFunctionResponse, typ, reason string) *ConditionOption {
	return newCondition(rsp, typ, reason, v1.Status_STATUS_CONDITION_UNKNOWN)
}

func newCondition(rsp *v1.RunFunctionResponse, typ, reason string, s v1.Status) *ConditionOption {
	if rsp.GetConditions() == nil {
		rsp.Conditions = make([]*v1.Condition, 0, 1)
	}
	c := &v1.Condition{
		Type:   typ,
		Status: s,
		Reason: reason,
		Target: v1.Target_TARGET_COMPOSITE.Enum(),
	}
	rsp.Conditions = append(rsp.GetConditions(), c)
	return &ConditionOption{condition: c}
}

// TargetComposite updates the condition to target the composite resource.
func (c *ConditionOption) TargetComposite() *ConditionOption {
	c.condition.Target = v1.Target_TARGET_COMPOSITE.Enum()
	return c
}

// TargetCompositeAndClaim updates the condition to target both the composite
// resource and claim.
func (c *ConditionOption) TargetCompositeAndClaim() *ConditionOption {
	c.condition.Target = v1.Target_TARGET_COMPOSITE_AND_CLAIM.Enum()
	return c
}

// WithMessage adds the message to the condition.
func (c *ConditionOption) WithMessage(message string) *ConditionOption {
	c.condition.Message = &message
	return c
}
