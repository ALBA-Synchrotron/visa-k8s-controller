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

package resource

import (
	"testing"

	cloudproviderv1 "git.cells.es/mis/visa/visa-k8s-controller/api/v1"
	"git.cells.es/mis/visa/visa-k8s-controller/internal/resolver"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPodBuilder(t *testing.T) {
	g := NewWithT(t)

	const testNs = "test-ns"

	builder := PodBuilder{}

	instance := &cloudproviderv1.VISAInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-instance",
			Namespace: testNs,
			Labels: map[string]string{
				"custom.label": "value1",
			},
			Annotations: map[string]string{
				"custom.annotation": "value2",
			},
		},
		Spec: cloudproviderv1.VISAInstanceSpec{
			Image:   corev1.LocalObjectReference{Name: "image-1"},
			Flavour: corev1.LocalObjectReference{Name: "flavour-1"},
			SecurityGroups: []corev1.LocalObjectReference{
				{Name: "sg-1"},
				{Name: "sg-2"},
			},
			DesiredState:     cloudproviderv1.DesiredStateActive,
			RebootGeneration: 3,
		},
		Status: cloudproviderv1.VISAInstanceStatus{
			ObservedRebootGeneration: 2,
		},
	}

	config := &resolver.ResolvedConfig{
		Image: &cloudproviderv1.VISACloudImage{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "image-1",
				Namespace: testNs,
			},
			Spec: cloudproviderv1.VISACloudImageSpec{
				HomeTemplatePath: "/home/template",
				UserShell:        "/bin/zsh",
				DNSPolicy:        corev1.DNSClusterFirst,
				ContainerSpec: cloudproviderv1.ContainerSpec{
					FullImageReference: "ubuntu:24.04",
					Command:            []string{"/init"},
					Args:               []string{"--start"},
					ImagePullPolicy:    corev1.PullAlways,
					Env: []corev1.EnvVar{
						{Name: "FOO", Value: "BAR"},
					},
					Ports: []corev1.ContainerPort{
						{ContainerPort: 8080},
					},
				},
			},
		},
		Flavour: &cloudproviderv1.VISACloudFlavour{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "flavour-1",
				Namespace: testNs,
			},
		},
		Sidecars: []*cloudproviderv1.VISASidecarConfig{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sidecar-a",
					Namespace: testNs,
				},
				Spec: cloudproviderv1.VISASidecarConfigSpec{
					ContainerSpec: cloudproviderv1.ContainerSpec{
						FullImageReference: "sidecar-a:1.0",
					},
				},
			},
		},
		SecurityGroups: []*cloudproviderv1.VISACloudSecurityGroup{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sg-1",
					Namespace: testNs,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sg-2",
					Namespace: testNs,
				},
			},
		},
	}

	imagePullSecrets := []corev1.LocalObjectReference{
		{Name: "regcred"},
	}

	pod := builder.Build(instance, config, imagePullSecrets)

	g.Expect(pod).NotTo(BeNil())
	g.Expect(pod.Name).To(Equal("test-instance-runtime"))
	g.Expect(pod.Namespace).To(Equal(testNs))
	g.Expect(pod.Labels).To(HaveKeyWithValue("visa.k8s.cells.es/managed", "true"))
	g.Expect(pod.Labels).To(HaveKeyWithValue("custom.label", "value1"))
	g.Expect(pod.Annotations).To(HaveKeyWithValue("custom.annotation", "value2"))
	g.Expect(pod.Annotations).To(HaveKeyWithValue("visa.k8s.cells.es/security-groups", "sg-1,sg-2"))
	g.Expect(pod.Annotations).To(HaveKeyWithValue("visa.k8s.cells.es/home-template-path", "/home/template"))
	g.Expect(pod.Annotations).To(HaveKeyWithValue("visa.k8s.cells.es/user-shell", "/bin/zsh"))
	g.Expect(pod.Annotations).To(HaveKeyWithValue("visa.k8s.cells.es/reboot-generation", "2"))
	g.Expect(pod.Annotations).To(HaveKey("cloudprovider.visa.k8s/spec-hash"))
	g.Expect(pod.Spec.DNSPolicy).To(Equal(corev1.DNSClusterFirst))
	g.Expect(pod.Spec.ImagePullSecrets).To(Equal(imagePullSecrets))
	g.Expect(pod.Spec.Containers).To(HaveLen(2))
	g.Expect(pod.Spec.Containers[0].Name).To(Equal("primary"))
	g.Expect(pod.Spec.Containers[0].Image).To(Equal("ubuntu:24.04"))
	g.Expect(pod.Spec.Containers[0].Command).To(Equal([]string{"/init"}))
	g.Expect(pod.Spec.Containers[0].Args).To(Equal([]string{"--start"}))
	g.Expect(pod.Spec.Containers[0].ImagePullPolicy).To(Equal(corev1.PullAlways))
	g.Expect(pod.Spec.Containers[1].Name).To(Equal("sidecar-a"))
	g.Expect(pod.Spec.Containers[1].Image).To(Equal("sidecar-a:1.0"))
}

func TestHashVISAInstanceSpec(t *testing.T) {
	g := NewWithT(t)

	spec1 := cloudproviderv1.VISAInstanceSpec{
		Image:            corev1.LocalObjectReference{Name: "img1"},
		Flavour:          corev1.LocalObjectReference{Name: "flav1"},
		DesiredState:     cloudproviderv1.DesiredStateActive,
		RebootGeneration: 0,
	}

	spec2 := cloudproviderv1.VISAInstanceSpec{
		Image:            corev1.LocalObjectReference{Name: "img1"},
		Flavour:          corev1.LocalObjectReference{Name: "flav1"},
		DesiredState:     cloudproviderv1.DesiredStateActive,
		RebootGeneration: 1,
	}

	hash1 := HashVISAInstanceSpec(spec1)
	hash2 := HashVISAInstanceSpec(spec2)

	g.Expect(hash1).NotTo(BeEmpty())
	g.Expect(hash2).NotTo(BeEmpty())
	g.Expect(hash1).NotTo(Equal(hash2))
}
