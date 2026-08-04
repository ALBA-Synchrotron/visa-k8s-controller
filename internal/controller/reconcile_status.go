package controller

import (
	"context"

	cloudproviderv1 "git.cells.es/mis/visa/visa-k8s-controller/api/v1"
	"git.cells.es/mis/visa/visa-k8s-controller/internal/resource"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *VISAInstanceReconciler) syncStatusFromPod(
	ctx context.Context,
	instance *cloudproviderv1.VISAInstance,
	pod *corev1.Pod,
) error {
	patch := client.MergeFrom(instance.DeepCopy())

	if pod.Status.Phase == corev1.PodRunning && pod.Status.PodIP != "" {
		instance.Status.Address = &pod.Status.PodIP
	} else {
		instance.Status.Address = nil
	}

	specHash := resource.HashVISAInstanceSpec(instance.Spec)

	if specHash == pod.Annotations["cloudprovider.visa.k8s/spec-hash"] &&
		pod.Status.Phase == corev1.PodRunning {
		instance.Status.ObservedRebootGeneration = instance.Spec.RebootGeneration
	}

	switch {
	case !instance.DeletionTimestamp.IsZero():
		instance.Status.State = cloudproviderv1.StateDeleting

	case instance.Spec.RebootGeneration > instance.Status.ObservedRebootGeneration:
		instance.Status.State = cloudproviderv1.StateRebooting

	case instance.Spec.DesiredState == cloudproviderv1.DesiredStateStopped:
		switch pod.Status.Phase {
		case corev1.PodPending, corev1.PodRunning:
			instance.Status.State = cloudproviderv1.StateStopping

		case corev1.PodSucceeded, corev1.PodFailed:
			instance.Status.State = cloudproviderv1.StateStopped

		default:
			instance.Status.State = cloudproviderv1.StateUnknown
		}

	case instance.Spec.DesiredState == cloudproviderv1.DesiredStateActive:
		switch pod.Status.Phase {
		case corev1.PodPending:
			instance.Status.State = cloudproviderv1.StateStarting

		case corev1.PodRunning:
			instance.Status.State = cloudproviderv1.StateActive

		case corev1.PodSucceeded, corev1.PodFailed:
			instance.Status.State = cloudproviderv1.StateError

		default:
			instance.Status.State = cloudproviderv1.StateUnknown
		}

	default:
		instance.Status.State = cloudproviderv1.StateUnknown
	}

	syncConditions(instance)

	return r.Status().Patch(ctx, instance, patch)
}

func (r *VISAInstanceReconciler) syncStatusWithoutPod(
	ctx context.Context,
	instance *cloudproviderv1.VISAInstance,
) error {
	patch := client.MergeFrom(instance.DeepCopy())

	instance.Status.Address = nil

	switch {
	case !instance.DeletionTimestamp.IsZero():
		instance.Status.State = cloudproviderv1.StateDeleting

	case instance.Spec.DesiredState == cloudproviderv1.DesiredStateStopped:
		instance.Status.State = cloudproviderv1.StateStopped

	case instance.Spec.DesiredState == cloudproviderv1.DesiredStateActive:
		instance.Status.State = cloudproviderv1.StateBuilding

	default:
		instance.Status.State = cloudproviderv1.StateUnknown
	}

	syncConditions(instance)

	return r.Status().Patch(ctx, instance, patch)
}

func syncConditions(instance *cloudproviderv1.VISAInstance) {
	ready := instance.Status.GetCondition(cloudproviderv1.Ready)
	provisioned := instance.Status.GetCondition(cloudproviderv1.Provisioned)

	switch instance.Status.State {

	case cloudproviderv1.StateBuilding:
		ready.Set(
			corev1.ConditionFalse,
			"PodCreating",
			"Waiting for backing pod to be created",
		)
		provisioned.Set(
			corev1.ConditionFalse,
			"PodCreating",
			"Backing pod has not been created",
		)

	case cloudproviderv1.StateStarting:
		ready.Set(
			corev1.ConditionFalse,
			"PodStarting",
			"Waiting for backing pod to become ready",
		)
		provisioned.Set(
			corev1.ConditionTrue,
			"PodCreated",
			"Backing pod has been created",
		)

	case cloudproviderv1.StateActive:
		ready.Set(
			corev1.ConditionTrue,
			"Ready",
			"Instance is ready",
		)
		provisioned.Set(
			corev1.ConditionTrue,
			"PodCreated",
			"Backing pod has been created",
		)

	case cloudproviderv1.StateStopping:
		ready.Set(
			corev1.ConditionFalse,
			"Stopping",
			"Instance is stopping",
		)
		provisioned.Set(
			corev1.ConditionTrue,
			"PodCreated",
			"Backing pod still exists",
		)

	case cloudproviderv1.StateStopped:
		ready.Set(
			corev1.ConditionFalse,
			"Stopped",
			"Instance is stopped",
		)
		provisioned.Set(
			corev1.ConditionFalse,
			"PodDeleted",
			"Backing pod does not exist",
		)

	case cloudproviderv1.StateRebooting:
		ready.Set(
			corev1.ConditionFalse,
			"Rebooting",
			"Instance is rebooting",
		)
		provisioned.Set(
			corev1.ConditionTrue,
			"PodCreated",
			"Backing pod exists",
		)

	case cloudproviderv1.StateDeleting:
		ready.Set(
			corev1.ConditionFalse,
			"Deleting",
			"Instance is being deleted",
		)
		provisioned.Set(
			corev1.ConditionFalse,
			"Deleting",
			"Backing pod is being deleted",
		)

	case cloudproviderv1.StateError:
		ready.Set(
			corev1.ConditionFalse,
			"PodFailed",
			"Backing pod has failed",
		)
		provisioned.Set(
			corev1.ConditionTrue,
			"PodCreated",
			"Backing pod exists but is unhealthy",
		)

	default:
		ready.Set(
			corev1.ConditionUnknown,
			"Unknown",
			"Unable to determine readiness",
		)
		provisioned.Set(
			corev1.ConditionUnknown,
			"Unknown",
			"Unable to determine provisioning state",
		)
	}
}
