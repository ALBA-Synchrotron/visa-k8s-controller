package status

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Ready            VISAInstanceConditionType = "Ready"
	NoWarnings       VISAInstanceConditionType = "NoWarnings"
	ReconcileSuccess VISAInstanceConditionType = "ReconcileSuccess"
)

type VISAInstanceConditionType string

// +kubebuilder:object:generate=true

type VISAInstanceCondition struct {
	Type VISAInstanceConditionType `json:"type"`
	// True, False, or Unknown
	Status corev1.ConditionStatus `json:"status"`
	// The last time this Condition type changed.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// One word, camel-case reason for current status of the condition.
	Reason string `json:"reason,omitempty"`
	// Full text reason for current status of the condition.
	Message string `json:"message,omitempty"`
}

func newVISAInstanceCondition(conditionType VISAInstanceConditionType) VISAInstanceCondition {
	return VISAInstanceCondition{
		Type:               conditionType,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
	}
}

func (c *VISAInstanceCondition) SetStatus(status corev1.ConditionStatus) {
	if c.Status != status {
		c.Status = status
		c.LastTransitionTime = metav1.Now()
	}
}

func (c *VISAInstanceCondition) Set(status corev1.ConditionStatus, reason, message string) {
	c.SetStatus(status)
	c.Reason = reason
	c.Message = message
}
