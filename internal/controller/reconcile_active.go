package controller

import (
	"context"

	cloudproviderv1 "git.cells.es/mis/visa/visa-k8s-controller/api/v1"
	"git.cells.es/mis/visa/visa-k8s-controller/internal/resolver"
	"git.cells.es/mis/visa/visa-k8s-controller/internal/resource"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *VISAInstanceReconciler) reconcileActive(ctx context.Context, visaInstance *cloudproviderv1.VISAInstance) (ctrl.Result, error) {

	resolvedConfig, err := resolver.New(r.Client).ResolveConfiguration(ctx, visaInstance)
	if err != nil {
		return ctrl.Result{}, err
	}

	desiredPod := resource.PodBuilder{}.Build(visaInstance, resolvedConfig)

	if err := controllerutil.SetControllerReference(
		visaInstance,
		desiredPod,
		r.Scheme,
	); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcilePod(ctx, desiredPod); err != nil {
		return ctrl.Result{}, err
	}

	actualPod, err := r.getPod(ctx, visaInstance)
	switch {
	case k8serrors.IsNotFound(err):
		return ctrl.Result{}, r.syncStatusWithoutPod(ctx, visaInstance)

	case err != nil:
		return ctrl.Result{}, err
	}

	if err := r.syncStatusFromPod(ctx, visaInstance, actualPod); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
