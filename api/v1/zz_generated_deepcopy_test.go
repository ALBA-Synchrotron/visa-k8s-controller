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
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeepCopyTypes(t *testing.T) {
	g := NewWithT(t)

	// ContainerSpec
	var nilCS *ContainerSpec
	g.Expect(nilCS.DeepCopy()).To(BeNil())

	cs := &ContainerSpec{
		FullImageReference: "img:1.0",
		Env: []corev1.EnvVar{
			{Name: "ENV1", Value: "VAL1"},
		},
		Ports: []corev1.ContainerPort{
			{ContainerPort: 8080},
		},
		Command: []string{"/bin/sh"},
		Args:    []string{"-c"},
	}
	csCopy := cs.DeepCopy()
	g.Expect(csCopy).To(Equal(cs))
	g.Expect(csCopy).NotTo(BeIdenticalTo(cs))

	// VISACloudDevice
	var nilDev *VISACloudDevice
	g.Expect(nilDev.DeepCopy()).To(BeNil())
	g.Expect(nilDev.DeepCopyObject()).To(BeNil())

	dev := &VISACloudDevice{
		ObjectMeta: metav1.ObjectMeta{Name: "dev1"},
		Spec:       VISACloudDeviceSpec{},
	}
	devCopy := dev.DeepCopy()
	g.Expect(devCopy).To(Equal(dev))
	g.Expect(dev.DeepCopyObject()).To(Equal(dev))

	var nilDevList *VISACloudDeviceList
	g.Expect(nilDevList.DeepCopy()).To(BeNil())
	g.Expect(nilDevList.DeepCopyObject()).To(BeNil())

	devList := &VISACloudDeviceList{
		Items: []VISACloudDevice{*dev},
	}
	g.Expect(devList.DeepCopy()).To(Equal(devList))
	g.Expect(devList.DeepCopyObject()).To(Equal(devList))

	var nilDevSpec *VISACloudDeviceSpec
	g.Expect(nilDevSpec.DeepCopy()).To(BeNil())
	devSpec := &VISACloudDeviceSpec{}
	g.Expect(devSpec.DeepCopy()).To(Equal(devSpec))

	// VISACloudFlavour
	var nilFlav *VISACloudFlavour
	g.Expect(nilFlav.DeepCopy()).To(BeNil())
	g.Expect(nilFlav.DeepCopyObject()).To(BeNil())

	flav := &VISACloudFlavour{
		ObjectMeta: metav1.ObjectMeta{Name: "flav1"},
		Spec: VISACloudFlavourSpec{
			Sidecars: []corev1.LocalObjectReference{{Name: "s1"}},
		},
	}
	g.Expect(flav.DeepCopy()).To(Equal(flav))
	g.Expect(flav.DeepCopyObject()).To(Equal(flav))

	var nilFlavList *VISACloudFlavourList
	g.Expect(nilFlavList.DeepCopy()).To(BeNil())
	g.Expect(nilFlavList.DeepCopyObject()).To(BeNil())

	flavList := &VISACloudFlavourList{
		Items: []VISACloudFlavour{*flav},
	}
	g.Expect(flavList.DeepCopy()).To(Equal(flavList))
	g.Expect(flavList.DeepCopyObject()).To(Equal(flavList))

	var nilFlavSpec *VISACloudFlavourSpec
	g.Expect(nilFlavSpec.DeepCopy()).To(BeNil())
	g.Expect(flav.Spec.DeepCopy()).To(Equal(&flav.Spec))

	// VISACloudImage
	var nilImg *VISACloudImage
	g.Expect(nilImg.DeepCopy()).To(BeNil())
	g.Expect(nilImg.DeepCopyObject()).To(BeNil())

	img := &VISACloudImage{
		ObjectMeta: metav1.ObjectMeta{Name: "img1"},
		Spec: VISACloudImageSpec{
			ContainerSpec:    *cs,
			HomeTemplatePath: "/home/tmpl",
			UserShell:        "/bin/bash",
			Size:             "1000",
		},
	}
	g.Expect(img.DeepCopy()).To(Equal(img))
	g.Expect(img.DeepCopyObject()).To(Equal(img))

	var nilImgList *VISACloudImageList
	g.Expect(nilImgList.DeepCopy()).To(BeNil())
	g.Expect(nilImgList.DeepCopyObject()).To(BeNil())

	imgList := &VISACloudImageList{
		Items: []VISACloudImage{*img},
	}
	g.Expect(imgList.DeepCopy()).To(Equal(imgList))
	g.Expect(imgList.DeepCopyObject()).To(Equal(imgList))

	var nilImgSpec *VISACloudImageSpec
	g.Expect(nilImgSpec.DeepCopy()).To(BeNil())
	g.Expect(img.Spec.DeepCopy()).To(Equal(&img.Spec))

	// VISACloudSecurityGroup
	var nilSG *VISACloudSecurityGroup
	g.Expect(nilSG.DeepCopy()).To(BeNil())
	g.Expect(nilSG.DeepCopyObject()).To(BeNil())

	sg := &VISACloudSecurityGroup{
		ObjectMeta: metav1.ObjectMeta{Name: "sg1"},
	}
	g.Expect(sg.DeepCopy()).To(Equal(sg))
	g.Expect(sg.DeepCopyObject()).To(Equal(sg))

	var nilSGList *VISACloudSecurityGroupList
	g.Expect(nilSGList.DeepCopy()).To(BeNil())
	g.Expect(nilSGList.DeepCopyObject()).To(BeNil())

	sgList := &VISACloudSecurityGroupList{
		Items: []VISACloudSecurityGroup{*sg},
	}
	g.Expect(sgList.DeepCopy()).To(Equal(sgList))
	g.Expect(sgList.DeepCopyObject()).To(Equal(sgList))

	// VISAInstance
	var nilInst *VISACloudInstance
	g.Expect(nilInst.DeepCopy()).To(BeNil())
	g.Expect(nilInst.DeepCopyObject()).To(BeNil())

	addr := "1.2.3.4"
	inst := &VISACloudInstance{
		ObjectMeta: metav1.ObjectMeta{Name: "inst1"},
		Spec: VISACloudInstanceSpec{
			Image:               corev1.LocalObjectReference{Name: "img1"},
			Flavour:             corev1.LocalObjectReference{Name: "flav1"},
			SecurityGroups:      []corev1.LocalObjectReference{{Name: "sg1"}},
			AssociatedProposals: []string{"prop1"},
		},
		Status: VISACloudInstanceStatus{
			Address: &addr,
			Conditions: []VISACloudInstanceCondition{
				{Type: Ready, Status: corev1.ConditionTrue},
			},
		},
	}
	g.Expect(inst.DeepCopy()).To(Equal(inst))
	g.Expect(inst.DeepCopyObject()).To(Equal(inst))

	var nilInstList *VISACloudInstanceList
	g.Expect(nilInstList.DeepCopy()).To(BeNil())
	g.Expect(nilInstList.DeepCopyObject()).To(BeNil())

	instList := &VISACloudInstanceList{
		Items: []VISACloudInstance{*inst},
	}
	g.Expect(instList.DeepCopy()).To(Equal(instList))
	g.Expect(instList.DeepCopyObject()).To(Equal(instList))

	var nilInstSpec *VISACloudInstanceSpec
	g.Expect(nilInstSpec.DeepCopy()).To(BeNil())
	g.Expect(inst.Spec.DeepCopy()).To(Equal(&inst.Spec))

	var nilInstStatus *VISACloudInstanceStatus
	g.Expect(nilInstStatus.DeepCopy()).To(BeNil())
	g.Expect(inst.Status.DeepCopy()).To(Equal(&inst.Status))

	var nilCond *VISACloudInstanceCondition
	g.Expect(nilCond.DeepCopy()).To(BeNil())
	cond := &VISACloudInstanceCondition{Type: Ready, Status: corev1.ConditionTrue}
	g.Expect(cond.DeepCopy()).To(Equal(cond))

	// VISASidecarConfig
	var nilSC *VISACloudSidecarConfig
	g.Expect(nilSC.DeepCopy()).To(BeNil())
	g.Expect(nilSC.DeepCopyObject()).To(BeNil())

	sc := &VISACloudSidecarConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "sc1"},
		Spec: VISACloudSidecarConfigSpec{
			ContainerSpec: *cs,
		},
	}
	g.Expect(sc.DeepCopy()).To(Equal(sc))
	g.Expect(sc.DeepCopyObject()).To(Equal(sc))

	var nilSCList *VISACloudSidecarConfigList
	g.Expect(nilSCList.DeepCopy()).To(BeNil())
	g.Expect(nilSCList.DeepCopyObject()).To(BeNil())

	scList := &VISACloudSidecarConfigList{
		Items: []VISACloudSidecarConfig{*sc},
	}
	g.Expect(scList.DeepCopy()).To(Equal(scList))
	g.Expect(scList.DeepCopyObject()).To(Equal(scList))

	var nilSCSpec *VISACloudSidecarConfigSpec
	g.Expect(nilSCSpec.DeepCopy()).To(BeNil())
	g.Expect(sc.Spec.DeepCopy()).To(Equal(&sc.Spec))
}
