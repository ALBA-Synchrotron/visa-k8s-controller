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

package resolver

import (
	"context"
	"testing"

	cloudproviderv1 "github.com/ALBA-Synchrotron/visa-k8s-controller/api/v1"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestConfigResolver(t *testing.T) {
	g := NewWithT(t)

	scheme := runtime.NewScheme()
	g.Expect(cloudproviderv1.AddToScheme(scheme)).To(Succeed())
	g.Expect(corev1.AddToScheme(scheme)).To(Succeed())

	ctx := context.Background()
	namespace := "test-ns"

	image := &cloudproviderv1.VISACloudImage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "img-1",
			Namespace: namespace,
		},
		Spec: cloudproviderv1.VISACloudImageSpec{
			ContainerSpec: cloudproviderv1.ContainerSpec{
				FullImageReference: "test-image:1.0",
			},
		},
	}

	flavour := &cloudproviderv1.VISACloudFlavour{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "flav-1",
			Namespace: namespace,
		},
		Spec: cloudproviderv1.VISACloudFlavourSpec{
			Sidecars: []corev1.LocalObjectReference{
				{Name: "sidecar-1"},
				{Name: "sidecar-2"},
			},
		},
	}

	sidecar1 := &cloudproviderv1.VISASidecarConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sidecar-1",
			Namespace: namespace,
		},
		Spec: cloudproviderv1.VISASidecarConfigSpec{
			ContainerSpec: cloudproviderv1.ContainerSpec{
				FullImageReference: "sidecar-img-1:latest",
			},
		},
	}

	sidecar2 := &cloudproviderv1.VISASidecarConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sidecar-2",
			Namespace: namespace,
		},
		Spec: cloudproviderv1.VISASidecarConfigSpec{
			ContainerSpec: cloudproviderv1.ContainerSpec{
				FullImageReference: "sidecar-img-2:latest",
			},
		},
	}

	sg1 := &cloudproviderv1.VISACloudSecurityGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sg-1",
			Namespace: namespace,
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(image, flavour, sidecar1, sidecar2, sg1).
		Build()

	r := New(fakeClient)

	instance := &cloudproviderv1.VISAInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "instance-1",
			Namespace: namespace,
		},
		Spec: cloudproviderv1.VISAInstanceSpec{
			Image:   corev1.LocalObjectReference{Name: "img-1"},
			Flavour: corev1.LocalObjectReference{Name: "flav-1"},
			SecurityGroups: []corev1.LocalObjectReference{
				{Name: "sg-1"},
			},
		},
	}

	// Successful resolution
	cfg, err := r.ResolveConfiguration(ctx, instance)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cfg).NotTo(BeNil())
	g.Expect(cfg.Image.Name).To(Equal("img-1"))
	g.Expect(cfg.Flavour.Name).To(Equal("flav-1"))
	g.Expect(cfg.Sidecars).To(HaveLen(2))
	g.Expect(cfg.SecurityGroups).To(HaveLen(1))

	// Missing Image error
	missingImgInstance := instance.DeepCopy()
	missingImgInstance.Spec.Image.Name = "non-existent-img"
	cfg, err = r.ResolveConfiguration(ctx, missingImgInstance)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("getting cloud image"))
	g.Expect(cfg).To(BeNil())

	// Missing Flavour error
	missingFlavInstance := instance.DeepCopy()
	missingFlavInstance.Spec.Flavour.Name = "non-existent-flav"
	cfg, err = r.ResolveConfiguration(ctx, missingFlavInstance)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("getting cloud flavour"))
	g.Expect(cfg).To(BeNil())

	// Missing Sidecar error
	missingSidecarFlav := flavour.DeepCopy()
	missingSidecarFlav.ResourceVersion = ""
	missingSidecarFlav.Name = "flav-missing-sidecar"
	missingSidecarFlav.Spec.Sidecars = []corev1.LocalObjectReference{{Name: "non-existent-sidecar"}}
	g.Expect(fakeClient.Create(ctx, missingSidecarFlav)).To(Succeed())

	missingSidecarInstance := instance.DeepCopy()
	missingSidecarInstance.Spec.Flavour.Name = "flav-missing-sidecar"
	cfg, err = r.ResolveConfiguration(ctx, missingSidecarInstance)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("getting cloud sidecar"))
	g.Expect(cfg).To(BeNil())

	// Missing Security Group error
	missingSGInstance := instance.DeepCopy()
	missingSGInstance.Spec.SecurityGroups = []corev1.LocalObjectReference{{Name: "non-existent-sg"}}
	cfg, err = r.ResolveConfiguration(ctx, missingSGInstance)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("getting cloud security group"))
	g.Expect(cfg).To(BeNil())
}
