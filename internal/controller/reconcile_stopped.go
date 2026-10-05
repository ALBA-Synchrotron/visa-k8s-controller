package controller

import (
	"context"

	cloudproviderv1 "github.com/ALBA-Synchrotron/visa-k8s-controller/api/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *VISAInstanceReconciler) reconcileStopped(ctx context.Context, instance *cloudproviderv1.VISAInstance) (ctrl.Result, error) {

	if err := r.deletePod(ctx, instance); err != nil {
		return ctrl.Result{}, err
	}

	pod, err := r.getPod(ctx, instance)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, r.syncStatusWithoutPod(ctx, instance)
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, r.syncStatusFromPod(ctx, instance, pod)
}
