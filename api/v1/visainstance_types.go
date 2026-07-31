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
	"git.cells.es/mis/visa/visa-k8s-controller/internal/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type VISAInstanceSpec struct {

	// +required
	Image LocalObjectReference `json:"image"`
	// +required
	Flavour LocalObjectReference `json:"flavour"`
	// +optional
	SecurityGroups []LocalObjectReference `json:"securityGroups,omitempty"`
	// +required
	// +kubebuilder:validation:Enum:=ACTIVE;DELETED;STOPPED
	DesiredState status.DesiredInstanceState `json:"desiredState"`
	// +optional
	RebootGeneration int64 `json:"rebootGeneration,omitempty"`
	// +optional
	AssociatedProposals []string `json:"associatedProposals,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
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
