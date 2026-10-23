package controller

import (
	"context"
	"fmt"
	"time"

	cloudproviderv1 "github.com/ALBA-Synchrotron/visa-k8s-controller/api/v1"
	"github.com/ALBA-Synchrotron/visa-k8s-controller/internal/resource"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *VISACloudInstanceReconciler) syncStatusFromPod(
	ctx context.Context,
	instance *cloudproviderv1.VISACloudInstance,
	pod *corev1.Pod,
) error {
	patch := client.MergeFrom(instance.DeepCopy())

	if pod.Status.Phase == corev1.PodRunning && pod.Status.PodIP != "" {
		instance.Status.Address = &pod.Status.PodIP
	} else {
		instance.Status.Address = nil
	}

	specHash := resource.HashVISACloudInstanceSpec(instance.Spec)

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

	instance.Status.Error = podErrorSummary(pod)
	syncConditions(instance)

	return r.Status().Patch(ctx, instance, patch)
}

func (r *VISACloudInstanceReconciler) syncStatusWithoutPod(
	ctx context.Context,
	instance *cloudproviderv1.VISACloudInstance,
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

func syncConditions(instance *cloudproviderv1.VISACloudInstance) {
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

func podErrorSummary(pod *corev1.Pod) cloudproviderv1.VISACloudInstanceStatusError {
	if err, ok := errorFromConditions(pod.Status.Conditions); ok {
		return err
	}

	if err, ok := errorFromContainerStatuses(pod); ok {
		return err
	}

	if err, ok := errorFromPodPhase(pod); ok {
		return err
	}

	return cloudproviderv1.VISACloudInstanceStatusError{}
}

func errorFromConditions(conditions []corev1.PodCondition) (cloudproviderv1.VISACloudInstanceStatusError, bool) {
	for _, condition := range conditions {
		if condition.Type == corev1.PodScheduled &&
			condition.Status == corev1.ConditionFalse &&
			(condition.Message != "" || condition.Reason != "") {
			createdAt := ""
			if !condition.LastTransitionTime.IsZero() {
				createdAt = condition.LastTransitionTime.UTC().Format(time.RFC3339)
			}
			code := condition.Reason
			if code == "" {
				code = string(condition.Type)
			}
			message := condition.Message
			if message == "" {
				message = condition.Reason
			}
			return cloudproviderv1.VISACloudInstanceStatusError{
				Code:      code,
				Message:   message,
				Details:   condition.Message,
				CreatedAt: createdAt,
			}, true
		}
	}
	return cloudproviderv1.VISACloudInstanceStatusError{}, false
}

func errorFromWaitingState(pod *corev1.Pod, waiting *corev1.ContainerStateWaiting) (cloudproviderv1.VISACloudInstanceStatusError, bool) {
	if waiting == nil {
		return cloudproviderv1.VISACloudInstanceStatusError{}, false
	}

	switch waiting.Reason {
	case "", "ContainerCreating", "PodInitializing":
		return cloudproviderv1.VISACloudInstanceStatusError{}, false
	}

	createdAt := ""
	if !pod.CreationTimestamp.IsZero() {
		createdAt = pod.CreationTimestamp.UTC().Format(time.RFC3339)
	}
	if pod.Status.StartTime != nil && !pod.Status.StartTime.IsZero() {
		createdAt = pod.Status.StartTime.UTC().Format(time.RFC3339)
	}

	message := waiting.Message
	if message == "" {
		message = waiting.Reason
	}

	return cloudproviderv1.VISACloudInstanceStatusError{
		Code:      waiting.Reason,
		Message:   message,
		Details:   waiting.Message,
		CreatedAt: createdAt,
	}, true
}

func errorFromTerminatedState(terminated *corev1.ContainerStateTerminated) (cloudproviderv1.VISACloudInstanceStatusError, bool) {
	if terminated == nil || terminated.Reason == "Completed" || (terminated.Reason == "" && terminated.ExitCode == 0 && terminated.Message == "") {
		return cloudproviderv1.VISACloudInstanceStatusError{}, false
	}

	createdAt := ""
	if !terminated.FinishedAt.IsZero() {
		createdAt = terminated.FinishedAt.UTC().Format(time.RFC3339)
	} else if !terminated.StartedAt.IsZero() {
		createdAt = terminated.StartedAt.UTC().Format(time.RFC3339)
	}

	code := terminated.Reason
	if code == "" {
		code = fmt.Sprintf("ExitCode%d", terminated.ExitCode)
	}

	message := terminated.Message
	if message == "" {
		if terminated.Reason != "" {
			message = terminated.Reason
		} else {
			message = fmt.Sprintf("Container terminated with exit code %d", terminated.ExitCode)
		}
	}

	return cloudproviderv1.VISACloudInstanceStatusError{
		Code:      code,
		Message:   message,
		Details:   terminated.Message,
		CreatedAt: createdAt,
	}, true
}

func errorFromContainerStatuses(pod *corev1.Pod) (cloudproviderv1.VISACloudInstanceStatusError, bool) {
	containerStatuses := append([]corev1.ContainerStatus{}, pod.Status.InitContainerStatuses...)
	containerStatuses = append(containerStatuses, pod.Status.ContainerStatuses...)

	for _, cs := range containerStatuses {
		if err, ok := errorFromWaitingState(pod, cs.State.Waiting); ok {
			return err, true
		}
		if err, ok := errorFromTerminatedState(cs.State.Terminated); ok {
			return err, true
		}
	}
	return cloudproviderv1.VISACloudInstanceStatusError{}, false
}

func errorFromPodPhase(pod *corev1.Pod) (cloudproviderv1.VISACloudInstanceStatusError, bool) {
	if pod.Status.Phase != corev1.PodFailed {
		return cloudproviderv1.VISACloudInstanceStatusError{}, false
	}

	createdAt := ""
	if !pod.CreationTimestamp.IsZero() {
		createdAt = pod.CreationTimestamp.UTC().Format(time.RFC3339)
	}
	if pod.Status.StartTime != nil && !pod.Status.StartTime.IsZero() {
		createdAt = pod.Status.StartTime.UTC().Format(time.RFC3339)
	}

	code := pod.Status.Reason
	if code == "" {
		code = "PodFailed"
	}

	message := pod.Status.Message
	if message == "" {
		if pod.Status.Reason != "" {
			message = pod.Status.Reason
		} else {
			message = "Pod failed"
		}
	}

	return cloudproviderv1.VISACloudInstanceStatusError{
		Code:      code,
		Message:   message,
		Details:   pod.Status.Message,
		CreatedAt: createdAt,
	}, true
}
