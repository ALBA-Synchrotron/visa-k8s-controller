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

const (
	errCodeUnknown = 0

	// Scheduling (1000s)
	errCodePodScheduling = 1000 // generic PodScheduled=False
	errCodeUnschedulable = 1001
	errCodeSchedulerErr  = 1002

	// Container waiting states (2000s)
	errCodeImagePull        = 2000 // ErrImagePull, ImagePullBackOff, ErrImageNeverPull, InvalidImageName
	errCodeCrashLoopBackOff = 2001
	errCodeContainerConfig  = 2002 // CreateContainerConfigError
	errCodeCreateContainer  = 2003 // CreateContainerError, RunContainerError
	errCodeWaitingOther     = 2099

	// Container terminated states (3000s)
	errCodeContainerError     = 3000 // Reason "Error"
	errCodeOOMKilled          = 3001
	errCodeContainerCannotRun = 3002
	errCodeDeadlineExceeded   = 3003
	errCodeContainerExitCode  = 3004 // no reason, non-zero exit code
	errCodeTerminatedOther    = 3099

	// Pod phase failures (4000s)
	errCodePodFailed       = 4000 // generic PodFailed
	errCodePodEvicted      = 4001
	errCodePodNodeLost     = 4002
	errCodePodShutdown     = 4003
	errCodePodAdmissionErr = 4004 // UnexpectedAdmissionError
	errCodePodFailedOther  = 4099
)

func (r *VISACloudInstanceReconciler) syncStatusFromPod(
	ctx context.Context,
	instance *cloudproviderv1.VISACloudInstance,
	pod *corev1.Pod,
) error {
	patch := client.MergeFrom(instance.DeepCopy())

	instance.Status.ObservedGeneration = instance.Generation

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

	instance.Status.ObservedGeneration = instance.Generation

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

func codeFromReason(reason string) int {
	switch reason {
	case corev1.PodReasonUnschedulable: // "Unschedulable"
		return errCodeUnschedulable
	case corev1.PodReasonSchedulerError: // "SchedulerError"
		return errCodeSchedulerErr
	case "":
		return errCodePodScheduling
	default:
		return errCodeUnknown
	}
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

			message := condition.Message
			if message == "" {
				message = condition.Reason
			}

			// Preserve the textual reason, since Code is now numeric.
			details := condition.Message
			if condition.Reason != "" {
				if details != "" {
					details = condition.Reason + ": " + details
				} else {
					details = condition.Reason
				}
			}

			return cloudproviderv1.VISACloudInstanceStatusError{
				Code:      codeFromReason(condition.Reason),
				Message:   message,
				Details:   details,
				CreatedAt: createdAt,
			}, true
		}
	}
	return cloudproviderv1.VISACloudInstanceStatusError{}, false
}

func codeFromWaitingReason(reason string) int {
	switch reason {
	case "ErrImagePull", "ImagePullBackOff", "ErrImageNeverPull", "InvalidImageName":
		return errCodeImagePull
	case "CrashLoopBackOff":
		return errCodeCrashLoopBackOff
	case "CreateContainerConfigError":
		return errCodeContainerConfig
	case "CreateContainerError", "RunContainerError":
		return errCodeCreateContainer
	default:
		return errCodeWaitingOther
	}
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
		Code:      codeFromWaitingReason(waiting.Reason),
		Message:   message,
		Details:   waiting.Message,
		CreatedAt: createdAt,
	}, true
}

func codeFromTerminatedState(terminated *corev1.ContainerStateTerminated) int {
	switch terminated.Reason {
	case "":
		return errCodeContainerExitCode
	case "Error":
		return errCodeContainerError
	case "OOMKilled":
		return errCodeOOMKilled
	case "ContainerCannotRun":
		return errCodeContainerCannotRun
	case "DeadlineExceeded":
		return errCodeDeadlineExceeded
	default:
		return errCodeTerminatedOther
	}
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

	message := terminated.Message
	if message == "" {
		if terminated.Reason != "" {
			message = terminated.Reason
		} else {
			message = fmt.Sprintf("Container terminated with exit code %d", terminated.ExitCode)
		}
	}

	return cloudproviderv1.VISACloudInstanceStatusError{
		Code:      codeFromTerminatedState(terminated),
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

func codeFromPodFailedReason(reason string) int {
	switch reason {
	case "":
		return errCodePodFailed
	case "Evicted":
		return errCodePodEvicted
	case "NodeLost":
		return errCodePodNodeLost
	case "Shutdown", "Terminated":
		return errCodePodShutdown
	case "UnexpectedAdmissionError":
		return errCodePodAdmissionErr
	default:
		return errCodePodFailedOther
	}
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

	message := pod.Status.Message
	if message == "" {
		if pod.Status.Reason != "" {
			message = pod.Status.Reason
		} else {
			message = "Pod failed"
		}
	}

	return cloudproviderv1.VISACloudInstanceStatusError{
		Code:      codeFromPodFailedReason(pod.Status.Reason),
		Message:   message,
		Details:   pod.Status.Message,
		CreatedAt: createdAt,
	}, true
}
