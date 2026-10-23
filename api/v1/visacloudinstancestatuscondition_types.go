package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Ready       VISACloudInstanceConditionType = "Ready"
	Provisioned VISACloudInstanceConditionType = "Provisioned"
)

type VISACloudInstanceConditionType string

// +kubebuilder:object:generate=true

type VISACloudInstanceCondition struct {
	// +required
	Type VISACloudInstanceConditionType `json:"type"`
	// +required
	Status corev1.ConditionStatus `json:"status"`
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// +optional
	Reason string `json:"reason,omitempty"`
	// +optional
	Message string `json:"message,omitempty"`
}

func newVISAInstanceCondition(t VISACloudInstanceConditionType) VISACloudInstanceCondition {
	return VISACloudInstanceCondition{
		Type:               t,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
	}
}

func (s *VISACloudInstanceStatus) GetCondition(
	t VISACloudInstanceConditionType,
) *VISACloudInstanceCondition {
	for i := range s.Conditions {
		if s.Conditions[i].Type == t {
			return &s.Conditions[i]
		}
	}

	s.Conditions = append(s.Conditions, newVISAInstanceCondition(t))
	return &s.Conditions[len(s.Conditions)-1]
}

func (c *VISACloudInstanceCondition) Set(
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
