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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type InstanceState string
type DesiredInstanceState string

const (
	StateActive          InstanceState = "ACTIVE"
	StatePartiallyActive InstanceState = "PARTIALLY_ACTIVE"
	StateBuilding        InstanceState = "BUILDING"
	StateDeleted         InstanceState = "DELETED"
	StateDeleting        InstanceState = "DELETING"
	StateError           InstanceState = "ERROR"
	StateStarting        InstanceState = "STARTING"
	StateStopped         InstanceState = "STOPPED"
	StateStopping        InstanceState = "STOPPING"
	StateRebooting       InstanceState = "REBOOTING"
	StateUnavailable     InstanceState = "UNAVAILABLE"
	StateUnknown         InstanceState = "UNKNOWN"
)

const (
	DesiredStateActive          DesiredInstanceState = "ACTIVE"
	DesiredStatePartiallyActive DesiredInstanceState = "PARTIALLY_ACTIVE"
	DesiredStateDeleted         DesiredInstanceState = "DELETED"
	DesiredStateStopped         DesiredInstanceState = "STOPPED"
)

type LocalObjectReference struct {
	// +required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
}

type VISAInstanceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// +required
	Image LocalObjectReference `json:"image"`
	// +required
	Flavour LocalObjectReference `json:"flavour"`
	// +optional
	SecurityGroups []string `json:"securityGroups,omitempty"`
	// +required
	// +kubebuilder:validation:Enum:=ACTIVE;PARTIALLY_ACTIVE;DELETED;STOPPED
	DesiredState DesiredInstanceState `json:"desiredState"`
	// +optional
	RebootGeneration int64 `json:"rebootGeneration,omitempty"`
	// +optional
	AssociatedProposals []string `json:"associatedProposals,omitempty"`
}
type VISAInstanceStatus struct {

	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// +required
	// +kubebuilder:validation:Enum:=ACTIVE;PARTIALLY_ACTIVE;BUILDING;DELETED;DELETING;ERROR;STARTING;STOPPED;STOPPING;REBOOTING;UNAVAILABLE;UNKNOWN
	State InstanceState `json:"state"`
	// +optional
	Address *string `json:"address,omitempty"`
	// +optional
	ObservedRebootGeneration int64 `json:"observedRebootGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type VISAInstance struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +required
	Spec VISAInstanceSpec `json:"spec"`
	// +optional
	Status VISAInstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type VISAInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VISAInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &VISAInstance{}, &VISAInstanceList{})
		return nil
	})
}
