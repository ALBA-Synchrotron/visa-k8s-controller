package v1

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = ginkgo.Describe("VISAInstance validation", func() {
	ginkgo.It("accepts valid object", func() {
		device := &VISAInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-visa-instance",
				Namespace: "default",
			},
			Spec: VISAInstanceSpec{
				Image: LocalObjectReference{
					Name: "test-image",
				},
				Flavour: LocalObjectReference{
					Name: "test-flavour",
				},
				DesiredState: DesiredStateActive,
			},
		}
		gomega.Expect(k8sClient.Create(ctx, device)).To(gomega.Succeed())
	})

	ginkgo.It("rejects empty image reference", func() {
		device := &VISAInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-visa-instance",
				Namespace: "default",
			},
			Spec: VISAInstanceSpec{
				Image: LocalObjectReference{
					Name: "",
				},
				Flavour: LocalObjectReference{
					Name: "test-flavour",
				},
				DesiredState: DesiredStateActive,
			},
		}
		err := k8sClient.Create(ctx, device)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("rejects invalid desired state", func() {
		device := &VISAInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-visa-instance",
				Namespace: "default",
			},
			Spec: VISAInstanceSpec{
				Image: LocalObjectReference{
					Name: "",
				},
				Flavour: LocalObjectReference{
					Name: "test-flavour",
				},
				DesiredState: "STARTING",
			},
		}
		err := k8sClient.Create(ctx, device)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
