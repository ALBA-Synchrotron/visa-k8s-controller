package controller

import (
	"context"

	cloudproviderv1 "git.cells.es/mis/visa/visa-k8s-controller/api/v1"
	clientretry "k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const deletionFinalizer = "deletion.finalizers.cloudprovider.visa.k8s"

// addFinalizerIfNeeded adds a deletion finalizer if the VISAInstance does not have one yet and is not marked for deletion
func (r *VISAInstanceReconciler) addFinalizerIfNeeded(ctx context.Context, visaInstance *cloudproviderv1.VISAInstance) error {
	if visaInstance.DeletionTimestamp.IsZero() && !controllerutil.ContainsFinalizer(visaInstance, deletionFinalizer) {
		return clientretry.RetryOnConflict(clientretry.DefaultRetry, func() error {
			controllerutil.AddFinalizer(visaInstance, deletionFinalizer)
			return r.Update(ctx, visaInstance)
		})
	}
	return nil
}

func (r *VISAInstanceReconciler) removeFinalizer(ctx context.Context, visaInstance *cloudproviderv1.VISAInstance) error {
	currentVisaInstance := &cloudproviderv1.VISAInstance{}
	currentVisaInstance.Name = visaInstance.Name
	currentVisaInstance.Namespace = visaInstance.Namespace

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, currentVisaInstance, func() error {
		controllerutil.RemoveFinalizer(currentVisaInstance, deletionFinalizer)
		return nil
	})

	if err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "Failed to remove finalizer for deletion")
		return client.IgnoreNotFound(err)
	}

	return nil
}
