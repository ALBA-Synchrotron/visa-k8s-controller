package v1

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
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
				Sidecars: []SidecarConfigRef{
					{Name: "guacd"}, {Name: "visa-print"},
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
				Sidecars: []SidecarConfigRef{
					{Name: "guacd"}, {Name: "visa-print"}, {Name: "guacd"},
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
				Sidecars: []SidecarConfigRef{
					{Name: "guacd"}, {Name: "visa-print"}, {Name: ""},
				},
			},
		}
		err := k8sClient.Create(ctx, device)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

})
