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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type VISASidecarConfigSpec struct {

	// +required
	// +kubebuilder:validation:MinLength=1
	FullImageReference string `json:"fullImageReference"`
	// +required
	Resources v1.ResourceRequirements `json:"resources"`
	// +optional
	Command []string `json:"command,omitempty"`
	// +optional
	Args []string `json:"args,omitempty"`
	// +optional
	Ports []v1.ContainerPort `json:"ports,omitempty"`
	// +required
	DeploySidecar bool `json:"deploySidecar"`
}

// +kubebuilder:object:root=true
type VISASidecarConfig struct {
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec VISASidecarConfigSpec `json:"spec"`
}

// +kubebuilder:object:root=true
type VISASidecarConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VISASidecarConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &VISASidecarConfig{}, &VISASidecarConfigList{})
		return nil
	})
}
