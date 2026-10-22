package v1

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = ginkgo.Describe("VISACloudFlavour validation", func() {
	ginkgo.It("accepts valid object", func() {
		device := &VISACloudFlavour{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-flavour",
				Namespace: "default",
			},
			Spec: VISACloudFlavourSpec{
				Sidecars: []corev1.LocalObjectReference{
					{Name: "guacd"}, {Name: "visa-print"},
				},
				DiskSize: "10Gi",
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
		}
		gomega.Expect(k8sClient.Create(ctx, device)).To(gomega.Succeed())
	})

	ginkgo.It("rejects duplicate sidecar references", func() {
		device := &VISACloudFlavour{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-flavour",
				Namespace: "default",
			},
			Spec: VISACloudFlavourSpec{
				Sidecars: []corev1.LocalObjectReference{
					{Name: "guacd"}, {Name: "visa-print"}, {Name: "guacd"},
				},
				DiskSize: "10Gi",
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
		}
		err := k8sClient.Create(ctx, device)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("rejects empty sidecar references", func() {
		device := &VISACloudFlavour{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-flavour",
				Namespace: "default",
			},
			Spec: VISACloudFlavourSpec{
				Sidecars: []corev1.LocalObjectReference{
					{Name: "guacd"}, {Name: "visa-print"}, {Name: ""},
				},
				DiskSize: "10Gi",
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
		}
		err := k8sClient.Create(ctx, device)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("rejects empty limits", func() {
		device := &VISACloudFlavour{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-flavour",
				Namespace: "default",
			},
			Spec: VISACloudFlavourSpec{
				Sidecars: []corev1.LocalObjectReference{
					{Name: "guacd"}, {Name: "visa-print"}, {Name: ""},
				},
				DiskSize: "10Gi",
				Limits:   corev1.ResourceList{},
			},
		}
		err := k8sClient.Create(ctx, device)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("rejects incomplete limits", func() {
		device := &VISACloudFlavour{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-flavour",
				Namespace: "default",
			},
			Spec: VISACloudFlavourSpec{
				Sidecars: []corev1.LocalObjectReference{
					{Name: "guacd"}, {Name: "visa-print"}, {Name: ""},
				},
				DiskSize: "10Gi",
				Limits: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("500m"),
				},
			},
		}
		err := k8sClient.Create(ctx, device)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("rejects missing diskSize", func() {
		device := &VISACloudFlavour{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-flavour",
				Namespace: "default",
			},
			Spec: VISACloudFlavourSpec{
				Sidecars: []corev1.LocalObjectReference{
					{Name: "guacd"}, {Name: "visa-print"}, {Name: ""},
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("500m"),
				},
			},
		}
		err := k8sClient.Create(ctx, device)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
