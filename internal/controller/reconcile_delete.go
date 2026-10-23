package controller

import (
	"context"

	cloudproviderv1 "github.com/ALBA-Synchrotron/visa-k8s-controller/api/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const deletionFinalizer = "cloudprovider.visa.k8s/finalizer"

func (r *VISACloudInstanceReconciler) ensureFinalizer(ctx context.Context, instance *cloudproviderv1.VISACloudInstance) error {
	if controllerutil.ContainsFinalizer(instance, deletionFinalizer) {
		return nil
	}
	controllerutil.AddFinalizer(instance, deletionFinalizer)
	return r.Update(ctx, instance)
}

func (r *VISACloudInstanceReconciler) removeFinalizer(ctx context.Context, instance *cloudproviderv1.VISACloudInstance) error {
	controllerutil.RemoveFinalizer(instance, deletionFinalizer)
	return client.IgnoreNotFound(
		r.Update(ctx, instance),
	)
}

func (r *VISACloudInstanceReconciler) reconcileDelete(ctx context.Context, instance *cloudproviderv1.VISACloudInstance) (ctrl.Result, error) {

	if !controllerutil.ContainsFinalizer(instance, deletionFinalizer) {
		return ctrl.Result{}, nil
	}

	if err := r.syncStatusWithoutPod(ctx, instance); err != nil {
		return ctrl.Result{}, err
	}

	_, err := r.deleteResources(ctx, instance)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := r.removeFinalizer(ctx, instance); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *VISACloudInstanceReconciler) deleteResources(ctx context.Context, instance *cloudproviderv1.VISACloudInstance) (bool, error) {

	pod, err := r.getPod(ctx, instance)

	logger := logf.FromContext(ctx)
	logger.Info("Deleting backing pod", "instance", instance.Name)

	switch {
	case k8serrors.IsNotFound(err):
		return true, nil

	case err != nil:
		return false, err
	}

	if err := r.Delete(ctx, pod); err != nil && !k8serrors.IsNotFound(err) {
		return false, err
	}

	return false, nil
}
