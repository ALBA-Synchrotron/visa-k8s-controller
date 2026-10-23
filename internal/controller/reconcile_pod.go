package controller

import (
	"context"

	cloudproviderv1 "github.com/ALBA-Synchrotron/visa-k8s-controller/api/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *VISACloudInstanceReconciler) reconcilePod(
	ctx context.Context,
	instance *cloudproviderv1.VISACloudInstance,
	desired *corev1.Pod,
) error {
	logger := logf.FromContext(ctx)
	current := &corev1.Pod{}

	err := r.Get(ctx, client.ObjectKey{
		Namespace: desired.Namespace,
		Name:      desired.Name,
	}, current)

	switch {
	case apierrors.IsNotFound(err):
		logger.Info("Creating backing pod of instance", "instance", instance.Name)
		if err := r.Create(ctx, desired); err != nil {
			r.Recorder.Eventf(
				instance,
				corev1.EventTypeWarning,
				"ProvisionFailed",
				"Failed to create backing pod %q: %v",
				desired.Name,
				err,
			)
			return err
		}

		r.Recorder.Eventf(
			instance,
			corev1.EventTypeNormal,
			"Provisioned",
			"Created backing pod %q",
			desired.Name,
		)

		return nil

	case err != nil:
		return err
	}

	desiredHash := desired.Annotations["cloudprovider.visa.k8s/spec-hash"]
	currentHash := current.Annotations["cloudprovider.visa.k8s/spec-hash"]

	if currentHash != desiredHash {
		if current.DeletionTimestamp.IsZero() {
			logger.Info("Reprovisioning (reboot) backing pod of instance", "instance", instance.Name)
			r.Recorder.Eventf(
				instance,
				corev1.EventTypeNormal,
				"Reprovisioning",
				"Backing pod %q is outdated and will be recreated",
				current.Name,
			)

			if err := r.Delete(ctx, current); err != nil {
				r.Recorder.Eventf(
					instance,
					corev1.EventTypeWarning,
					"ReprovisionFailed",
					"Failed to delete backing pod %q: %v",
					current.Name,
					err,
				)
				return err
			}
		}
	}

	return nil
}

func (r *VISACloudInstanceReconciler) deletePod(ctx context.Context, instance *cloudproviderv1.VISACloudInstance) error {
	pod := &corev1.Pod{}

	err := r.Get(ctx, client.ObjectKey{
		Namespace: instance.Namespace,
		Name:      instance.Name + "-runtime",
	}, pod)

	if pod.DeletionTimestamp.IsZero() {

		switch {
		case apierrors.IsNotFound(err):
			return nil

		case err != nil:
			r.Recorder.Eventf(
				instance,
				corev1.EventTypeWarning,
				"DeprovisionFailed",
				"Failed to retrieve backing pod: %v",
				err,
			)
			return err
		}

		if err := r.Delete(ctx, pod); err != nil {
			r.Recorder.Eventf(
				instance,
				corev1.EventTypeWarning,
				"DeprovisionFailed",
				"Failed to delete backing pod %q: %v",
				pod.Name,
				err,
			)
			return err
		}

		r.Recorder.Eventf(
			instance,
			corev1.EventTypeNormal,
			"Deprovisioned",
			"Deleted backing pod %q",
			pod.Name,
		)
	}

	return nil
}

func (r *VISACloudInstanceReconciler) getPod(ctx context.Context, instance *cloudproviderv1.VISACloudInstance) (*corev1.Pod, error) {

	pod := &corev1.Pod{}

	err := r.Get(ctx, client.ObjectKey{
		Namespace: instance.Namespace,
		Name:      instance.Name + "-runtime",
	}, pod)
	if err != nil {
		return nil, err
	}

	return pod, nil
}
