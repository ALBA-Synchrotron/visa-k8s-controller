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

package controller

import (
	"context"

	cloudproviderv1 "git.cells.es/mis/visa/visa-k8s-controller/api/v1"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = ginkgo.Describe("VISAInstance Controller", func() {
	ginkgo.Context("When reconciling a resource", func() {
		const (
			resourceName      = "test-visa-instance"
			resourceNamespace = "default"
		)

		ctx := context.Background()

		var (
			visaInstance         *cloudproviderv1.VISAInstance
			image                *cloudproviderv1.VISACloudImage
			flavour              *cloudproviderv1.VISACloudFlavour
			jupyterSidecarConfig *cloudproviderv1.VISASidecarConfig
			securityGroup        *cloudproviderv1.VISACloudSecurityGroup
			controllerReconciler *VISAInstanceReconciler
		)

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: resourceNamespace,
		}

		ginkgo.BeforeEach(func() {
			ginkgo.By("creating the required resources")

			image = &cloudproviderv1.VISACloudImage{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-image",
					Namespace: resourceNamespace,
				},
				Spec: cloudproviderv1.VISACloudImageSpec{
					Size:             "2654.23",
					HomeTemplatePath: "/etc/home/template",
					UserShell:        "/usr/bin/bash",
					ContainerSpec: cloudproviderv1.ContainerSpec{
						FullImageReference: "docker.io/busybox:1.37",
					},
				},
			}
			gomega.Expect(k8sClient.Create(ctx, image)).Should(gomega.Succeed())

			flavour = &cloudproviderv1.VISACloudFlavour{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-flavour",
					Namespace: resourceNamespace,
				},
				Spec: cloudproviderv1.VISACloudFlavourSpec{
					Sidecars: []corev1.LocalObjectReference{
						{Name: "guacd"},
					},
				},
			}
			gomega.Expect(k8sClient.Create(ctx, flavour)).Should(gomega.Succeed())

			jupyterSidecarConfig = &cloudproviderv1.VISASidecarConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "guacd",
					Namespace: resourceNamespace,
				},
				Spec: cloudproviderv1.VISASidecarConfigSpec{
					ContainerSpec: cloudproviderv1.ContainerSpec{
						FullImageReference: "docker.io/guacamole/guacd:latest",
					},
				},
			}
			gomega.Expect(k8sClient.Create(ctx, jupyterSidecarConfig)).Should(gomega.Succeed())

			securityGroup = &cloudproviderv1.VISACloudSecurityGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sg",
					Namespace: resourceNamespace,
				},
			}
			gomega.Expect(k8sClient.Create(ctx, securityGroup)).Should(gomega.Succeed())

			visaInstance = &cloudproviderv1.VISAInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: resourceNamespace,
				},
				Spec: cloudproviderv1.VISAInstanceSpec{
					Image: corev1.LocalObjectReference{
						Name: "test-image",
					},
					Flavour: corev1.LocalObjectReference{
						Name: "test-flavour",
					},
					SecurityGroups: []corev1.LocalObjectReference{
						{Name: "test-sg"},
					},
					DesiredState: cloudproviderv1.DesiredStateActive,
				},
			}
			gomega.Expect(k8sClient.Create(ctx, visaInstance)).Should(gomega.Succeed())

			controllerReconciler = &VISAInstanceReconciler{
				Client:           k8sClient,
				Scheme:           k8sClient.Scheme(),
				Recorder:         record.NewFakeRecorder(100),
				ImagePullSecrets: []corev1.LocalObjectReference{{Name: "test-secret"}},
			}
		})

		ginkgo.AfterEach(func() {
			ginkgo.By("cleanup of created fixtures")

			if visaInstance != nil {
				fetchedInstance := &cloudproviderv1.VISAInstance{}
				if err := k8sClient.Get(ctx, typeNamespacedName, fetchedInstance); err == nil {
					fetchedInstance.Finalizers = nil
					_ = k8sClient.Update(ctx, fetchedInstance)
					_ = k8sClient.Delete(ctx, fetchedInstance)
				}
			}

			pod := &corev1.Pod{}
			if err := k8sClient.Get(ctx, types.NamespacedName{Namespace: resourceNamespace, Name: resourceName + "-runtime"}, pod); err == nil {
				_ = k8sClient.Delete(ctx, pod)
			}

			gomega.Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, securityGroup))).To(gomega.Succeed())
			gomega.Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, jupyterSidecarConfig))).To(gomega.Succeed())
			gomega.Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, flavour))).To(gomega.Succeed())
			gomega.Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, image))).To(gomega.Succeed())
		})

		ginkgo.It("should successfully reconcile the resource", func() {
			ginkgo.By("Reconciling the created resource")
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			ginkgo.By("verifying the backing pod was created with expected configuration")
			createdPod := &corev1.Pod{}
			podKey := types.NamespacedName{
				Namespace: resourceNamespace,
				Name:      resourceName + "-runtime",
			}
			gomega.Expect(k8sClient.Get(ctx, podKey, createdPod)).To(gomega.Succeed())
			gomega.Expect(createdPod.Labels).To(gomega.HaveKeyWithValue("visa.k8s.cells.es/managed", "true"))
			gomega.Expect(createdPod.Annotations).To(gomega.HaveKeyWithValue("visa.k8s.cells.es/home-template-path", "/etc/home/template"))
			gomega.Expect(createdPod.Annotations).To(gomega.HaveKeyWithValue("visa.k8s.cells.es/user-shell", "/usr/bin/bash"))
			gomega.Expect(createdPod.Annotations).To(gomega.HaveKeyWithValue("visa.k8s.cells.es/security-groups", "test-sg"))
			gomega.Expect(createdPod.Spec.Containers).To(gomega.HaveLen(2))
			gomega.Expect(createdPod.Spec.Containers[0].Name).To(gomega.Equal("primary"))
			gomega.Expect(createdPod.Spec.Containers[0].Image).To(gomega.Equal("docker.io/busybox:1.37"))
			gomega.Expect(createdPod.Spec.Containers[1].Name).To(gomega.Equal("guacd"))
			gomega.Expect(createdPod.Spec.Containers[1].Image).To(gomega.Equal("docker.io/guacamole/guacd:latest"))
			gomega.Expect(createdPod.OwnerReferences).To(gomega.HaveLen(1))
			gomega.Expect(createdPod.OwnerReferences[0].Name).To(gomega.Equal(resourceName))

			ginkgo.By("verifying the instance has the finalizer added")
			updatedInstance := &cloudproviderv1.VISAInstance{}
			gomega.Expect(k8sClient.Get(ctx, typeNamespacedName, updatedInstance)).To(gomega.Succeed())
			gomega.Expect(updatedInstance.Finalizers).To(gomega.ContainElement("cloudprovider.visa.k8s/finalizer"))

			ginkgo.By("verifying status transition when pod is Pending")
			createdPod.Status.Phase = corev1.PodPending
			gomega.Expect(k8sClient.Status().Update(ctx, createdPod)).To(gomega.Succeed())

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(k8sClient.Get(ctx, typeNamespacedName, updatedInstance)).To(gomega.Succeed())
			gomega.Expect(updatedInstance.Status.State).To(gomega.Equal(cloudproviderv1.StateStarting))
			provisionedCond := updatedInstance.Status.GetCondition(cloudproviderv1.Provisioned)
			gomega.Expect(provisionedCond.Status).To(gomega.Equal(corev1.ConditionTrue))
			readyCond := updatedInstance.Status.GetCondition(cloudproviderv1.Ready)
			gomega.Expect(readyCond.Status).To(gomega.Equal(corev1.ConditionFalse))

			ginkgo.By("verifying status transition when pod is Running with IP")
			createdPod.Status.Phase = corev1.PodRunning
			createdPod.Status.PodIP = "10.244.0.5"
			gomega.Expect(k8sClient.Status().Update(ctx, createdPod)).To(gomega.Succeed())

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(k8sClient.Get(ctx, typeNamespacedName, updatedInstance)).To(gomega.Succeed())
			gomega.Expect(updatedInstance.Status.State).To(gomega.Equal(cloudproviderv1.StateActive))
			gomega.Expect(updatedInstance.Status.Address).NotTo(gomega.BeNil())
			gomega.Expect(*updatedInstance.Status.Address).To(gomega.Equal("10.244.0.5"))
			readyCond = updatedInstance.Status.GetCondition(cloudproviderv1.Ready)
			gomega.Expect(readyCond.Status).To(gomega.Equal(corev1.ConditionTrue))
			provisionedCond = updatedInstance.Status.GetCondition(cloudproviderv1.Provisioned)
			gomega.Expect(provisionedCond.Status).To(gomega.Equal(corev1.ConditionTrue))
		})

		ginkgo.It("should stop the instance when desiredState is STOPPED", func() {
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			podKey := types.NamespacedName{
				Namespace: resourceNamespace,
				Name:      resourceName + "-runtime",
			}
			createdPod := &corev1.Pod{}
			gomega.Expect(k8sClient.Get(ctx, podKey, createdPod)).To(gomega.Succeed())

			ginkgo.By("changing desiredState to STOPPED")
			updatedInstance := &cloudproviderv1.VISAInstance{}
			gomega.Expect(k8sClient.Get(ctx, typeNamespacedName, updatedInstance)).To(gomega.Succeed())
			updatedInstance.Spec.DesiredState = cloudproviderv1.DesiredStateStopped
			gomega.Expect(k8sClient.Update(ctx, updatedInstance)).To(gomega.Succeed())

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			ginkgo.By("verifying backing pod is deleted")
			podErr := k8sClient.Get(ctx, podKey, createdPod)
			gomega.Expect(k8serrors.IsNotFound(podErr)).To(gomega.BeTrue())

			ginkgo.By("verifying instance state is STOPPED")
			gomega.Expect(k8sClient.Get(ctx, typeNamespacedName, updatedInstance)).To(gomega.Succeed())
			gomega.Expect(updatedInstance.Status.State).To(gomega.Equal(cloudproviderv1.StateStopped))
			readyCond := updatedInstance.Status.GetCondition(cloudproviderv1.Ready)
			gomega.Expect(readyCond.Status).To(gomega.Equal(corev1.ConditionFalse))
			provisionedCond := updatedInstance.Status.GetCondition(cloudproviderv1.Provisioned)
			gomega.Expect(provisionedCond.Status).To(gomega.Equal(corev1.ConditionFalse))
		})

		ginkgo.It("should restart a stopped instance when desiredState changes back to ACTIVE", func() {
			podKey := types.NamespacedName{
				Namespace: resourceNamespace,
				Name:      resourceName + "-runtime",
			}

			ginkgo.By("1. Reconciling to ACTIVE initially")
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			createdPod := &corev1.Pod{}
			gomega.Expect(k8sClient.Get(ctx, podKey, createdPod)).To(gomega.Succeed())

			ginkgo.By("2. Transitioning instance to STOPPED")
			inst := &cloudproviderv1.VISAInstance{}
			gomega.Expect(k8sClient.Get(ctx, typeNamespacedName, inst)).To(gomega.Succeed())
			inst.Spec.DesiredState = cloudproviderv1.DesiredStateStopped
			gomega.Expect(k8sClient.Update(ctx, inst)).To(gomega.Succeed())

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(k8serrors.IsNotFound(k8sClient.Get(ctx, podKey, createdPod))).To(gomega.BeTrue())
			gomega.Expect(k8sClient.Get(ctx, typeNamespacedName, inst)).To(gomega.Succeed())
			gomega.Expect(inst.Status.State).To(gomega.Equal(cloudproviderv1.StateStopped))

			ginkgo.By("3. Changing desiredState back to ACTIVE (Restart)")
			inst.Spec.DesiredState = cloudproviderv1.DesiredStateActive
			gomega.Expect(k8sClient.Update(ctx, inst)).To(gomega.Succeed())

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			ginkgo.By("4. Verifying new backing pod is created")
			restartedPod := &corev1.Pod{}
			gomega.Expect(k8sClient.Get(ctx, podKey, restartedPod)).To(gomega.Succeed())

			ginkgo.By("5. Simulating PodRunning phase with PodIP")
			restartedPod.Status.Phase = corev1.PodRunning
			restartedPod.Status.PodIP = "10.244.0.42"
			gomega.Expect(k8sClient.Status().Update(ctx, restartedPod)).To(gomega.Succeed())

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(k8sClient.Get(ctx, typeNamespacedName, inst)).To(gomega.Succeed())
			gomega.Expect(inst.Status.State).To(gomega.Equal(cloudproviderv1.StateActive))
			gomega.Expect(inst.Status.Address).NotTo(gomega.BeNil())
			gomega.Expect(*inst.Status.Address).To(gomega.Equal("10.244.0.42"))
			gomega.Expect(inst.Status.GetCondition(cloudproviderv1.Ready).Status).To(gomega.Equal(corev1.ConditionTrue))
			gomega.Expect(inst.Status.GetCondition(cloudproviderv1.Provisioned).Status).To(gomega.Equal(corev1.ConditionTrue))
		})

		ginkgo.It("should handle instance reboot when rebootGeneration is incremented", func() {
			podKey := types.NamespacedName{
				Namespace: resourceNamespace,
				Name:      resourceName + "-runtime",
			}

			ginkgo.By("1. Initial reconcile to ACTIVE with running pod")
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			initialPod := &corev1.Pod{}
			gomega.Expect(k8sClient.Get(ctx, podKey, initialPod)).To(gomega.Succeed())
			initialPod.Status.Phase = corev1.PodRunning
			initialPod.Status.PodIP = "10.244.0.10"
			gomega.Expect(k8sClient.Status().Update(ctx, initialPod)).To(gomega.Succeed())

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			inst := &cloudproviderv1.VISAInstance{}
			gomega.Expect(k8sClient.Get(ctx, typeNamespacedName, inst)).To(gomega.Succeed())
			gomega.Expect(inst.Status.State).To(gomega.Equal(cloudproviderv1.StateActive))
			gomega.Expect(inst.Status.ObservedRebootGeneration).To(gomega.Equal(int64(0)))

			ginkgo.By("2. Incrementing rebootGeneration to 1")
			inst.Spec.RebootGeneration = 1
			gomega.Expect(k8sClient.Update(ctx, inst)).To(gomega.Succeed())

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			ginkgo.By("3. Verifying outdated pod was deleted and state is BUILDING")
			gomega.Expect(k8serrors.IsNotFound(k8sClient.Get(ctx, podKey, initialPod))).To(gomega.BeTrue())

			gomega.Expect(k8sClient.Get(ctx, typeNamespacedName, inst)).To(gomega.Succeed())
			gomega.Expect(inst.Status.State).To(gomega.Equal(cloudproviderv1.StateBuilding))

			ginkgo.By("4. Next reconcile creates the new rebooted pod and enters REBOOTING while pod is pending")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			newPod := &corev1.Pod{}
			gomega.Expect(k8sClient.Get(ctx, podKey, newPod)).To(gomega.Succeed())
			newPod.Status.Phase = corev1.PodPending
			gomega.Expect(k8sClient.Status().Update(ctx, newPod)).To(gomega.Succeed())

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(k8sClient.Get(ctx, typeNamespacedName, inst)).To(gomega.Succeed())
			gomega.Expect(inst.Status.State).To(gomega.Equal(cloudproviderv1.StateRebooting))
			gomega.Expect(inst.Status.GetCondition(cloudproviderv1.Ready).Status).To(gomega.Equal(corev1.ConditionFalse))
			gomega.Expect(inst.Status.GetCondition(cloudproviderv1.Ready).Reason).To(gomega.Equal("Rebooting"))

			ginkgo.By("5. Reconcile with PodRunning updates observedRebootGeneration and returns to StateActive")
			newPod.Status.Phase = corev1.PodRunning
			newPod.Status.PodIP = "10.244.0.11"
			gomega.Expect(k8sClient.Status().Update(ctx, newPod)).To(gomega.Succeed())

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(k8sClient.Get(ctx, typeNamespacedName, inst)).To(gomega.Succeed())
			gomega.Expect(inst.Status.ObservedRebootGeneration).To(gomega.Equal(int64(1)))
			gomega.Expect(inst.Status.State).To(gomega.Equal(cloudproviderv1.StateActive))
			gomega.Expect(inst.Status.GetCondition(cloudproviderv1.Ready).Status).To(gomega.Equal(corev1.ConditionTrue))
		})

		ginkgo.It("should pause reconciliation when pauseReconciliation label is set", func() {
			ginkgo.By("setting pause reconciliation label")
			inst := &cloudproviderv1.VISAInstance{}
			gomega.Expect(k8sClient.Get(ctx, typeNamespacedName, inst)).To(gomega.Succeed())
			if inst.Labels == nil {
				inst.Labels = make(map[string]string)
			}
			inst.Labels[pauseReconciliationLabel] = "true"
			gomega.Expect(k8sClient.Update(ctx, inst)).To(gomega.Succeed())

			res, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(res).To(gomega.Equal(ctrl.Result{}))

			podKey := types.NamespacedName{
				Namespace: resourceNamespace,
				Name:      resourceName + "-runtime",
			}
			pod := &corev1.Pod{}
			gomega.Expect(k8serrors.IsNotFound(k8sClient.Get(ctx, podKey, pod))).To(gomega.BeTrue())
		})

		ginkgo.It("should return error when referenced resources are missing", func() {
			missingInst := &cloudproviderv1.VISAInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "missing-res-instance",
					Namespace: resourceNamespace,
				},
				Spec: cloudproviderv1.VISAInstanceSpec{
					Image:        corev1.LocalObjectReference{Name: "non-existent-img"},
					Flavour:      corev1.LocalObjectReference{Name: "test-flavour"},
					DesiredState: cloudproviderv1.DesiredStateActive,
				},
			}
			gomega.Expect(k8sClient.Create(ctx, missingInst)).To(gomega.Succeed())
			defer func() {
				_ = k8sClient.Delete(ctx, missingInst)
			}()

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: resourceNamespace,
					Name:      "missing-res-instance",
				},
			})
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("getting cloud image"))
		})

		ginkgo.It("should handle various pod states and error reporting", func() {
			podKey := types.NamespacedName{
				Namespace: resourceNamespace,
				Name:      resourceName + "-runtime",
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			pod := &corev1.Pod{}
			gomega.Expect(k8sClient.Get(ctx, podKey, pod)).To(gomega.Succeed())

			// 1. PodScheduled=False with message
			pod.Status.Conditions = []corev1.PodCondition{
				{
					Type:    corev1.PodScheduled,
					Status:  corev1.ConditionFalse,
					Message: "0/1 nodes available: insufficient memory",
				},
			}
			gomega.Expect(k8sClient.Status().Update(ctx, pod)).To(gomega.Succeed())

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			inst := &cloudproviderv1.VISAInstance{}
			gomega.Expect(k8sClient.Get(ctx, typeNamespacedName, inst)).To(gomega.Succeed())
			gomega.Expect(inst.Status.Error).To(gomega.Equal("0/1 nodes available: insufficient memory"))

			// 2. Container waiting state
			pod.Status.Conditions = nil
			pod.Status.ContainerStatuses = []corev1.ContainerStatus{
				{
					Name: "primary",
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason:  "ImagePullBackOff",
							Message: "Back-off pulling image",
						},
					},
				},
			}
			gomega.Expect(k8sClient.Status().Update(ctx, pod)).To(gomega.Succeed())

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(k8sClient.Get(ctx, typeNamespacedName, inst)).To(gomega.Succeed())
			gomega.Expect(inst.Status.Error).To(gomega.Equal("ImagePullBackOff: Back-off pulling image"))

			// 3. Container terminated state
			pod.Status.ContainerStatuses = []corev1.ContainerStatus{
				{
					Name: "primary",
					State: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{
							Reason:  "OOMKilled",
							Message: "Container killed due to OOM",
						},
					},
				},
			}
			gomega.Expect(k8sClient.Status().Update(ctx, pod)).To(gomega.Succeed())

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(k8sClient.Get(ctx, typeNamespacedName, inst)).To(gomega.Succeed())
			gomega.Expect(inst.Status.Error).To(gomega.Equal("OOMKilled: Container killed due to OOM"))

			// 4. Pod in PodFailed phase
			pod.Status.ContainerStatuses = nil
			pod.Status.Phase = corev1.PodFailed
			pod.Status.Message = "Node was evicted"
			gomega.Expect(k8sClient.Status().Update(ctx, pod)).To(gomega.Succeed())

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(k8sClient.Get(ctx, typeNamespacedName, inst)).To(gomega.Succeed())
			gomega.Expect(inst.Status.State).To(gomega.Equal(cloudproviderv1.StateError))
			gomega.Expect(inst.Status.Error).To(gomega.Equal("Node was evicted"))

			// 5. Pod in PodFailed phase with Reason only
			pod.Status.Message = ""
			pod.Status.Reason = "Evicted"
			gomega.Expect(k8sClient.Status().Update(ctx, pod)).To(gomega.Succeed())

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(k8sClient.Get(ctx, typeNamespacedName, inst)).To(gomega.Succeed())
			gomega.Expect(inst.Status.Error).To(gomega.Equal("Evicted"))

			// 6. Pod in PodFailed phase without reason/message
			pod.Status.Reason = ""
			gomega.Expect(k8sClient.Status().Update(ctx, pod)).To(gomega.Succeed())

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(k8sClient.Get(ctx, typeNamespacedName, inst)).To(gomega.Succeed())
			gomega.Expect(inst.Status.Error).To(gomega.Equal("Pod failed"))

			// 7. Pod in PodSucceeded phase
			pod.Status.Phase = corev1.PodSucceeded
			gomega.Expect(k8sClient.Status().Update(ctx, pod)).To(gomega.Succeed())

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(k8sClient.Get(ctx, typeNamespacedName, inst)).To(gomega.Succeed())
			gomega.Expect(inst.Status.State).To(gomega.Equal(cloudproviderv1.StateError))

			// 8. DesiredState STOPPED with Pod still Pending (StateStopping)
			inst.Spec.DesiredState = cloudproviderv1.DesiredStateStopped
			gomega.Expect(k8sClient.Update(ctx, inst)).To(gomega.Succeed())

			pod.Status.Phase = corev1.PodPending
			_ = controllerReconciler.syncStatusFromPod(ctx, inst, pod)
			gomega.Expect(inst.Status.State).To(gomega.Equal(cloudproviderv1.StateStopping))

			// 9. DesiredState STOPPED with Pod Succeeded/Failed (StateStopped)
			pod.Status.Phase = corev1.PodSucceeded
			_ = controllerReconciler.syncStatusFromPod(ctx, inst, pod)
			gomega.Expect(inst.Status.State).To(gomega.Equal(cloudproviderv1.StateStopped))

			// 10. DesiredState unknown
			inst.Spec.DesiredState = "UNKNOWN_STATE"
			_ = controllerReconciler.syncStatusFromPod(ctx, inst, pod)
			gomega.Expect(inst.Status.State).To(gomega.Equal(cloudproviderv1.StateUnknown))

			inst.Spec.DesiredState = "UNKNOWN_STATE"
			_ = controllerReconciler.syncStatusWithoutPod(ctx, inst)
			gomega.Expect(inst.Status.State).To(gomega.Equal(cloudproviderv1.StateUnknown))

			// 11. Container waiting with reason only (empty message)
			pod.Status.Phase = corev1.PodPending
			pod.Status.ContainerStatuses = []corev1.ContainerStatus{
				{
					Name: "primary",
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason: "CrashLoopBackOff",
						},
					},
				},
			}
			inst.Spec.DesiredState = cloudproviderv1.DesiredStateActive
			_ = controllerReconciler.syncStatusFromPod(ctx, inst, pod)
			gomega.Expect(inst.Status.Error).To(gomega.Equal("CrashLoopBackOff"))

			// 12. Container waiting with ignored reasons
			pod.Status.ContainerStatuses = []corev1.ContainerStatus{
				{
					Name: "primary",
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason: "ContainerCreating",
						},
					},
				},
			}
			_ = controllerReconciler.syncStatusFromPod(ctx, inst, pod)
			gomega.Expect(inst.Status.Error).To(gomega.Equal(""))

			// 13. Container terminated with reason only (empty message)
			pod.Status.ContainerStatuses = []corev1.ContainerStatus{
				{
					Name: "primary",
					State: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{
							Reason: "Error",
						},
					},
				},
			}
			_ = controllerReconciler.syncStatusFromPod(ctx, inst, pod)
			gomega.Expect(inst.Status.Error).To(gomega.Equal("Error"))

			// 14. PodScheduled condition false without message
			pod.Status.ContainerStatuses = nil
			pod.Status.Conditions = []corev1.PodCondition{
				{
					Type:   corev1.PodScheduled,
					Status: corev1.ConditionFalse,
				},
			}
			_ = controllerReconciler.syncStatusFromPod(ctx, inst, pod)
			gomega.Expect(inst.Status.Error).To(gomega.Equal(""))

			// 15. Pod running without PodIP
			pod.Status.Phase = corev1.PodRunning
			pod.Status.PodIP = ""
			_ = controllerReconciler.syncStatusFromPod(ctx, inst, pod)
			gomega.Expect(inst.Status.Address).To(gomega.BeNil())

			// 16. Instance with DeletionTimestamp in syncStatusFromPod and syncStatusWithoutPod
			now := metav1.Now()
			inst.DeletionTimestamp = &now
			_ = controllerReconciler.syncStatusFromPod(ctx, inst, pod)
			gomega.Expect(inst.Status.State).To(gomega.Equal(cloudproviderv1.StateDeleting))

			inst.DeletionTimestamp = &now
			_ = controllerReconciler.syncStatusWithoutPod(ctx, inst)
			gomega.Expect(inst.Status.State).To(gomega.Equal(cloudproviderv1.StateDeleting))
			inst.DeletionTimestamp = nil
		})

		ginkgo.It("should clean up resources and remove finalizer when deleting", func() {
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			podKey := types.NamespacedName{
				Namespace: resourceNamespace,
				Name:      resourceName + "-runtime",
			}
			createdPod := &corev1.Pod{}
			gomega.Expect(k8sClient.Get(ctx, podKey, createdPod)).To(gomega.Succeed())

			ginkgo.By("deleting the instance")
			instanceToDelete := &cloudproviderv1.VISAInstance{}
			gomega.Expect(k8sClient.Get(ctx, typeNamespacedName, instanceToDelete)).To(gomega.Succeed())
			gomega.Expect(k8sClient.Delete(ctx, instanceToDelete)).To(gomega.Succeed())

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			ginkgo.By("verifying backing pod is deleted")
			podErr := k8sClient.Get(ctx, podKey, createdPod)
			gomega.Expect(k8serrors.IsNotFound(podErr)).To(gomega.BeTrue())

			ginkgo.By("verifying instance is deleted completely")
			instanceErr := k8sClient.Get(ctx, typeNamespacedName, instanceToDelete)
			gomega.Expect(k8serrors.IsNotFound(instanceErr)).To(gomega.BeTrue())
		})

		ginkgo.It("should handle non-existent resource gracefully", func() {
			res, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: resourceNamespace,
					Name:      "non-existent-instance-xyz",
				},
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(res).To(gomega.Equal(ctrl.Result{}))
		})

		ginkgo.It("should handle reconcileDelete when finalizer is already removed", func() {
			inst := &cloudproviderv1.VISAInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-finalizer-inst",
					Namespace: resourceNamespace,
				},
			}
			res, err := controllerReconciler.reconcileDelete(ctx, inst)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(res).To(gomega.Equal(ctrl.Result{}))
		})

		ginkgo.It("should setup with manager", func() {
			mgr, err := ctrl.NewManager(cfg, ctrl.Options{
				Scheme: k8sClient.Scheme(),
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			r := &VISAInstanceReconciler{
				Client:   mgr.GetClient(),
				Scheme:   mgr.GetScheme(),
				Recorder: mgr.GetEventRecorderFor("visainstance-controller"),
			}
			err = r.SetupWithManager(mgr)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.It("should handle owner reference failure in reconcileActive", func() {
			badReconciler := &VISAInstanceReconciler{
				Client:   k8sClient,
				Scheme:   runtime.NewScheme(),
				Recorder: record.NewFakeRecorder(10),
			}
			_, err := badReconciler.reconcileActive(ctx, visaInstance)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("no kind is registered"))
		})

		ginkgo.It("should handle deletePod when pod is already marked for deletion", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deleting-pod-runtime",
					Namespace: resourceNamespace,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "test", Image: "busybox"}},
				},
			}
			gomega.Expect(k8sClient.Create(ctx, pod)).To(gomega.Succeed())
			defer func() {
				_ = k8sClient.Delete(ctx, pod)
			}()

			now := metav1.Now()
			pod.DeletionTimestamp = &now
			testInst := &cloudproviderv1.VISAInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deleting-pod",
					Namespace: resourceNamespace,
				},
			}
			err := controllerReconciler.deletePod(ctx, testInst)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.It("should handle error branches with injected errors", func() {
			testErr := k8serrors.NewInternalError(context.DeadlineExceeded)

			// 1. getVISAInstance error in Reconcile
			errCl := &errorClient{Client: k8sClient, getErr: testErr}
			r := &VISAInstanceReconciler{Client: errCl, Scheme: k8sClient.Scheme(), Recorder: record.NewFakeRecorder(10)}
			_, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			gomega.Expect(err).To(gomega.HaveOccurred())

			// 2. ensureFinalizer error in Reconcile
			instWithoutFinalizer := visaInstance.DeepCopy()
			instWithoutFinalizer.Finalizers = nil
			updateErrCl := &errorClient{Client: k8sClient, updateErr: testErr}
			r = &VISAInstanceReconciler{Client: updateErrCl, Scheme: k8sClient.Scheme(), Recorder: record.NewFakeRecorder(10)}
			err = r.ensureFinalizer(ctx, instWithoutFinalizer)
			gomega.Expect(err).To(gomega.HaveOccurred())

			// 3. Pause reconciliation with Status Patch error
			if visaInstance.Labels == nil {
				visaInstance.Labels = make(map[string]string)
			}
			visaInstance.Labels[pauseReconciliationLabel] = "true"
			gomega.Expect(k8sClient.Update(ctx, visaInstance)).To(gomega.Succeed())

			statusErrCl := &errorStatusClient{Client: k8sClient, patchErr: testErr}
			r = &VISAInstanceReconciler{Client: statusErrCl, Scheme: k8sClient.Scheme(), Recorder: record.NewFakeRecorder(10)}
			res, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(res).To(gomega.Equal(ctrl.Result{}))

			delete(visaInstance.Labels, pauseReconciliationLabel)
			gomega.Expect(k8sClient.Update(ctx, visaInstance)).To(gomega.Succeed())

			// 4. reconcilePod Create error
			createErrCl := &errorClient{Client: k8sClient, createErr: testErr}
			r = &VISAInstanceReconciler{Client: createErrCl, Scheme: k8sClient.Scheme(), Recorder: record.NewFakeRecorder(10)}
			desiredPod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "non-existent-pod-xyz", Namespace: resourceNamespace}}
			err = r.reconcilePod(ctx, visaInstance, desiredPod)
			gomega.Expect(err).To(gomega.HaveOccurred())

			// 5. reconcilePod Delete error (during reprovision)
			podForReprovision := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "pod-for-reprovision",
					Namespace:   resourceNamespace,
					Annotations: map[string]string{"cloudprovider.visa.k8s/spec-hash": "old-hash"},
				},
				Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "t", Image: "busybox"}}},
			}
			gomega.Expect(k8sClient.Create(ctx, podForReprovision)).To(gomega.Succeed())
			defer func() { _ = k8sClient.Delete(ctx, podForReprovision) }()

			deleteErrCl := &errorClient{Client: k8sClient, deleteErr: testErr}
			r = &VISAInstanceReconciler{Client: deleteErrCl, Scheme: k8sClient.Scheme(), Recorder: record.NewFakeRecorder(10)}
			desiredReprovisionPod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "pod-for-reprovision",
					Namespace:   resourceNamespace,
					Annotations: map[string]string{"cloudprovider.visa.k8s/spec-hash": "new-hash"},
				},
			}
			err = r.reconcilePod(ctx, visaInstance, desiredReprovisionPod)
			gomega.Expect(err).To(gomega.HaveOccurred())

			// 6. reconcilePod Get error (other than NotFound)
			r = &VISAInstanceReconciler{Client: errCl, Scheme: k8sClient.Scheme(), Recorder: record.NewFakeRecorder(10)}
			err = r.reconcilePod(ctx, visaInstance, desiredReprovisionPod)
			gomega.Expect(err).To(gomega.HaveOccurred())

			// 7. deletePod Get error (other than NotFound)
			err = r.deletePod(ctx, visaInstance)
			gomega.Expect(err).To(gomega.HaveOccurred())

			// 8. deletePod Delete error
			r = &VISAInstanceReconciler{Client: deleteErrCl, Scheme: k8sClient.Scheme(), Recorder: record.NewFakeRecorder(10)}
			existingPod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      visaInstance.Name + "-runtime",
					Namespace: resourceNamespace,
				},
				Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "t", Image: "busybox"}}},
			}
			_ = k8sClient.Create(ctx, existingPod)
			err = r.deletePod(ctx, visaInstance)
			gomega.Expect(err).To(gomega.HaveOccurred())

			// 9. deleteResources Get error and Delete error
			r = &VISAInstanceReconciler{Client: errCl, Scheme: k8sClient.Scheme(), Recorder: record.NewFakeRecorder(10)}
			_, err = r.deleteResources(ctx, visaInstance)
			gomega.Expect(err).To(gomega.HaveOccurred())

			r = &VISAInstanceReconciler{Client: deleteErrCl, Scheme: k8sClient.Scheme(), Recorder: record.NewFakeRecorder(10)}
			_, err = r.deleteResources(ctx, visaInstance)
			gomega.Expect(err).To(gomega.HaveOccurred())

			// 10. reconcileDelete with errors
			r = &VISAInstanceReconciler{Client: statusErrCl, Scheme: k8sClient.Scheme(), Recorder: record.NewFakeRecorder(10)}
			instWithFinalizer := visaInstance.DeepCopy()
			instWithFinalizer.Finalizers = []string{deletionFinalizer}
			now := metav1.Now()
			instWithFinalizer.DeletionTimestamp = &now
			_, err = r.reconcileDelete(ctx, instWithFinalizer)
			gomega.Expect(err).To(gomega.HaveOccurred())

			r = &VISAInstanceReconciler{Client: errCl, Scheme: k8sClient.Scheme(), Recorder: record.NewFakeRecorder(10)}
			_, err = r.reconcileDelete(ctx, instWithFinalizer)
			gomega.Expect(err).To(gomega.HaveOccurred())

			r = &VISAInstanceReconciler{Client: updateErrCl, Scheme: k8sClient.Scheme(), Recorder: record.NewFakeRecorder(10)}
			_, err = r.reconcileDelete(ctx, instWithFinalizer)
			gomega.Expect(err).To(gomega.HaveOccurred())

			// 11. reconcileStopped error on deletePod
			r = &VISAInstanceReconciler{Client: errCl, Scheme: k8sClient.Scheme(), Recorder: record.NewFakeRecorder(10)}
			_, err = r.reconcileStopped(ctx, visaInstance)
			gomega.Expect(err).To(gomega.HaveOccurred())

			// 12. reconcileActive errors on getPod and syncStatusFromPod
			r = &VISAInstanceReconciler{Client: statusErrCl, Scheme: k8sClient.Scheme(), Recorder: record.NewFakeRecorder(10)}
			_, err = r.reconcileActive(ctx, visaInstance)
			gomega.Expect(err).To(gomega.HaveOccurred())

			// 13. Reconcile error when desiredState is STOPPED and deletePod fails
			stoppedInst := visaInstance.DeepCopy()
			stoppedInst.Spec.DesiredState = cloudproviderv1.DesiredStateStopped
			_ = k8sClient.Update(ctx, stoppedInst)
			r = &VISAInstanceReconciler{Client: errCl, Scheme: k8sClient.Scheme(), Recorder: record.NewFakeRecorder(10)}
			_, err = r.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
	})
})

type errorClient struct {
	client.Client
	getErr    error
	createErr error
	deleteErr error
	updateErr error
}

func (e *errorClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if e.getErr != nil {
		return e.getErr
	}
	return e.Client.Get(ctx, key, obj, opts...)
}

func (e *errorClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if e.createErr != nil {
		return e.createErr
	}
	return e.Client.Create(ctx, obj, opts...)
}

func (e *errorClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	if e.deleteErr != nil {
		return e.deleteErr
	}
	return e.Client.Delete(ctx, obj, opts...)
}

func (e *errorClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if e.updateErr != nil {
		return e.updateErr
	}
	return e.Client.Update(ctx, obj, opts...)
}

type errorStatusClient struct {
	client.Client
	patchErr error
}

func (e *errorStatusClient) Status() client.SubResourceWriter {
	return &errorStatusWriter{SubResourceWriter: e.Client.Status(), patchErr: e.patchErr}
}

type errorStatusWriter struct {
	client.SubResourceWriter
	patchErr error
}

func (w *errorStatusWriter) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
	if w.patchErr != nil {
		return w.patchErr
	}
	return w.SubResourceWriter.Patch(ctx, obj, patch, opts...)
}
