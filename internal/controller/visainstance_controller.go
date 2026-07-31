/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	cloudproviderv1 "git.cells.es/mis/visa/visa-k8s-controller/api/v1"
	"git.cells.es/mis/visa/visa-k8s-controller/internal/resolver"
	"git.cells.es/mis/visa/visa-k8s-controller/internal/resource"
	"git.cells.es/mis/visa/visa-k8s-controller/internal/status"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// VISAInstanceReconciler reconciles a VISAInstance object
type VISAInstanceReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

const (
	pauseReconciliationLabel = "cloudprovider.visa.k8s/pauseReconciliation"
)

// +kubebuilder:rbac:groups=cloudprovider.visa.k8s,resources=visainstances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudprovider.visa.k8s,resources=visainstances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloudprovider.visa.k8s,resources=visainstances/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VISAInstance object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.24.1/pkg/reconcile
func (r *VISAInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	logger.Info("Reconciling VISAInstance")

	visaInstance, err := r.getVISAInstance(ctx, req.NamespacedName)

	if client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, err
	} else if k8serrors.IsNotFound(err) {
		// No need to requeue if the resource no longer exists
		return ctrl.Result{}, nil
	}

	// Create or update deployment based on desired state
	if visaInstance.Spec.DesiredState == status.DesiredStateActive {
		logger.Info("Creating or updating deployment for VISAInstance")

		// Update status to BUILDING while deployment is being created
		patch := client.MergeFrom(visaInstance.DeepCopy())
		if visaInstance.Status.State != status.StateActive {
			visaInstance.Status.State = status.StateBuilding
		}

		visaInstance, err = r.buildVISAInstance(ctx, visaInstance)

		visaInstance.Status.SetCondition(status.ReconcileSuccess, corev1.ConditionTrue, "reconciling")
		if err := r.Status().Patch(ctx, visaInstance, patch); err != nil {
			logger.Error(err, "Failed to update status")
			return ctrl.Result{}, err
		}
	} else if visaInstance.Spec.DesiredState == status.DesiredStateStopped {
		// TODO: Implement stop logic (scale deployment to 0 replicas)
		logger.Info("Stopping VISAInstance")
		patch := client.MergeFrom(visaInstance.DeepCopy())
		visaInstance.Status.State = status.StateStopping
		if err := r.Status().Patch(ctx, visaInstance, patch); err != nil {
			logger.Error(err, "Failed to update status to STOPPING")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VISAInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudproviderv1.VISAInstance{}).
		Named("visainstance").
		Complete(r)
}

func (r *VISAInstanceReconciler) getVISAInstance(ctx context.Context, namespacedName types.NamespacedName) (*cloudproviderv1.VISAInstance, error) {
	visaInstance := &cloudproviderv1.VISAInstance{}
	err := r.Get(ctx, namespacedName, visaInstance)
	return visaInstance, err
}

func (r *VISAInstanceReconciler) buildVISAInstance(ctx context.Context, visaInstance *cloudproviderv1.VISAInstance) (*cloudproviderv1.VISAInstance, error) {

	configResolver := resolver.New(r.Client)
	resolvedConfig, err := configResolver.ResolveConfiguration(ctx, visaInstance)
	if err != nil {
		return nil, err
	}

	podBuilder := resource.PodBuilder{}
	pod := podBuilder.Build(visaInstance, resolvedConfig)

	// TODO: Move out from here?
	err = r.Client.Create(ctx, pod)
	if err != nil {
		return nil, err
	}

	return visaInstance, nil
}
