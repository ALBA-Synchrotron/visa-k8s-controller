/*
Copyright 2026.

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

package v1

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

func TestVISACloudInstanceStatusConditions(t *testing.T) {
	g := NewWithT(t)

	status := &VISACloudInstanceStatus{}

	// Test GetCondition when condition does not exist
	cond := status.GetCondition(Ready)
	g.Expect(cond).NotTo(BeNil())
	g.Expect(cond.Type).To(Equal(Ready))
	g.Expect(cond.Status).To(Equal(corev1.ConditionUnknown))
	g.Expect(cond.LastTransitionTime.IsZero()).To(BeFalse())
	g.Expect(status.Conditions).To(HaveLen(1))

	// Test GetCondition when condition already exists
	condExisting := status.GetCondition(Ready)
	g.Expect(condExisting).To(Equal(cond))
	g.Expect(status.Conditions).To(HaveLen(1))

	// Test Set when status changes
	initialTransitionTime := cond.LastTransitionTime
	time.Sleep(10 * time.Millisecond)
	cond.Set(corev1.ConditionTrue, "InstanceReady", "Instance is running and ready")
	g.Expect(cond.Status).To(Equal(corev1.ConditionTrue))
	g.Expect(cond.Reason).To(Equal("InstanceReady"))
	g.Expect(cond.Message).To(Equal("Instance is running and ready"))
	g.Expect(cond.LastTransitionTime.After(initialTransitionTime.Time)).To(BeTrue())

	// Test Set when status is unchanged (LastTransitionTime should not change)
	secondTransitionTime := cond.LastTransitionTime
	cond.Set(corev1.ConditionTrue, "InstanceStillReady", "Updated message")
	g.Expect(cond.Status).To(Equal(corev1.ConditionTrue))
	g.Expect(cond.Reason).To(Equal("InstanceStillReady"))
	g.Expect(cond.Message).To(Equal("Updated message"))
	g.Expect(cond.LastTransitionTime).To(Equal(secondTransitionTime))

	// Test Provisioned condition
	provCond := status.GetCondition(Provisioned)
	g.Expect(provCond.Type).To(Equal(Provisioned))
	g.Expect(status.Conditions).To(HaveLen(2))
	provCond.Set(corev1.ConditionFalse, "NotProvisioned", "Provisioning in progress")
	g.Expect(provCond.Status).To(Equal(corev1.ConditionFalse))

	// Test newVISAInstanceCondition directly
	c := newVISAInstanceCondition(Ready)
	g.Expect(c.Type).To(Equal(Ready))
	g.Expect(c.Status).To(Equal(corev1.ConditionUnknown))
	g.Expect(c.LastTransitionTime.IsZero()).To(BeFalse())
}
