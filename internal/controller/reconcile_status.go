package controller

import (
	"context"

	cloudproviderv1 "git.cells.es/mis/visa/visa-k8s-controller/api/v1"
	"git.cells.es/mis/visa/visa-k8s-controller/internal/status"
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

	if instance.Spec.RebootGeneration > instance.Status.ObservedRebootGeneration &&
		pod.Status.Phase == corev1.PodRunning {
		instance.Status.ObservedRebootGeneration = instance.Spec.RebootGeneration
	}

	switch {

	case !instance.DeletionTimestamp.IsZero():
		instance.Status.State = status.StateDeleting

	case instance.Spec.RebootGeneration > instance.Status.ObservedRebootGeneration:
		instance.Status.State = status.StateRebooting

	case instance.Spec.DesiredState == status.DesiredStateStopped:
		switch pod.Status.Phase {

		case corev1.PodPending, corev1.PodRunning:
			instance.Status.State = status.StateStopping

		case corev1.PodSucceeded, corev1.PodFailed:
			instance.Status.State = status.StateStopped

		default:
			instance.Status.State = status.StateUnknown
		}

	case instance.Spec.DesiredState == status.DesiredStateActive:
		switch pod.Status.Phase {

		case corev1.PodPending:
			instance.Status.State = status.StateStarting

		case corev1.PodRunning:
			instance.Status.State = status.StateActive

		case corev1.PodSucceeded:
			instance.Status.State = status.StateError

		case corev1.PodFailed:
			instance.Status.State = status.StateError

		default:
			instance.Status.State = status.StateUnknown
		}

	default:
		instance.Status.State = status.StateUnknown
	}

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
		instance.Status.State = status.StateDeleting

	case instance.Spec.DesiredState == status.DesiredStateStopped:
		instance.Status.State = status.StateStopped

	case instance.Spec.DesiredState == status.DesiredStateActive:
		instance.Status.State = status.StateBuilding

	default:
		instance.Status.State = status.StateUnknown
	}

	return r.Status().Patch(ctx, instance, patch)
}
