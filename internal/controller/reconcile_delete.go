package controller

import (
	"context"

	cloudproviderv1 "git.cells.es/mis/visa/visa-k8s-controller/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const deletionFinalizer = "deletion.finalizers.cloudprovider.visa.k8s"

func (r *VISAInstanceReconciler) ensureFinalizer(ctx context.Context, instance *cloudproviderv1.VISAInstance) error {
	if controllerutil.ContainsFinalizer(instance, deletionFinalizer) {
		return nil
	}
	controllerutil.AddFinalizer(instance, deletionFinalizer)
	return r.Update(ctx, instance)
}

func (r *VISAInstanceReconciler) removeFinalizer(ctx context.Context, instance *cloudproviderv1.VISAInstance) error {
	controllerutil.RemoveFinalizer(instance, deletionFinalizer)
	return client.IgnoreNotFound(
		r.Update(ctx, instance),
	)
}

func (r *VISAInstanceReconciler) reconcileDelete(ctx context.Context, instance *cloudproviderv1.VISAInstance) (ctrl.Result, error) {

	if !controllerutil.ContainsFinalizer(instance, deletionFinalizer) {
		return ctrl.Result{}, nil
	}

	if err := r.deleteResources(ctx, instance); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.removeFinalizer(ctx, instance); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
