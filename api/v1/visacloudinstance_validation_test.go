package v1

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = ginkgo.Describe("VISACloudInstance validation", func() {
	ginkgo.It("accepts valid object", func() {
		device := &VISACloudInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-visa-instance",
				Namespace: "default",
			},
			Spec: VISACloudInstanceSpec{
				Image: corev1.LocalObjectReference{
					Name: "test-image",
				},
				Flavour: corev1.LocalObjectReference{
					Name: "test-flavour",
				},
				DesiredState: DesiredStateActive,
			},
		}
		gomega.Expect(k8sClient.Create(ctx, device)).To(gomega.Succeed())
	})

	ginkgo.It("rejects empty image reference", func() {
		device := &VISACloudInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-visa-instance",
				Namespace: "default",
			},
			Spec: VISACloudInstanceSpec{
				Image: corev1.LocalObjectReference{
					Name: "",
				},
				Flavour: corev1.LocalObjectReference{
					Name: "test-flavour",
				},
				DesiredState: DesiredStateActive,
			},
		}
		err := k8sClient.Create(ctx, device)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("rejects invalid desired state", func() {
		device := &VISACloudInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-visa-instance",
				Namespace: "default",
			},
			Spec: VISACloudInstanceSpec{
				Image: corev1.LocalObjectReference{
					Name: "",
				},
				Flavour: corev1.LocalObjectReference{
					Name: "test-flavour",
				},
				DesiredState: "STARTING",
			},
		}
		err := k8sClient.Create(ctx, device)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
