package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Ready       VISAInstanceConditionType = "Ready"
	Provisioned VISAInstanceConditionType = "Provisioned"
)

type VISAInstanceConditionType string

// +kubebuilder:object:generate=true

type VISAInstanceCondition struct {
	// +required
	Type VISAInstanceConditionType `json:"type"`
	// +required
	Status corev1.ConditionStatus `json:"status"`
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// +optional
	Reason string `json:"reason,omitempty"`
	// +optional
	Message string `json:"message,omitempty"`
}

func newVISAInstanceCondition(t VISAInstanceConditionType) VISAInstanceCondition {
	return VISAInstanceCondition{
		Type:               t,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
	}
}

func (s *VISAInstanceStatus) GetCondition(
	t VISAInstanceConditionType,
) *VISAInstanceCondition {
	for i := range s.Conditions {
		if s.Conditions[i].Type == t {
			return &s.Conditions[i]
		}
	}

	s.Conditions = append(s.Conditions, newVISAInstanceCondition(t))
	return &s.Conditions[len(s.Conditions)-1]
}

func (c *VISAInstanceCondition) Set(
	status corev1.ConditionStatus,
	reason string,
	message string,
) {
	if c.Status != status {
		c.Status = status
		c.LastTransitionTime = metav1.Now()
	}

	c.Reason = reason
	c.Message = message
}
