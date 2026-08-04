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
	"fmt"

	cloudproviderv1 "git.cells.es/mis/visa/visa-k8s-controller/api/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

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

func (r *VISAInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var (
		result ctrl.Result
		err    error
	)

	logger := logf.FromContext(ctx)

	logger.Info("Reconciling VISAInstance")

	visaInstance, err := r.getVISAInstance(ctx, req.NamespacedName)

	if client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, err
	} else if k8serrors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}

	if v, ok := visaInstance.Labels[pauseReconciliationLabel]; ok && v == "true" {
		logger.Info("Not reconciling VISAInstance")
		r.Recorder.Event(visaInstance, corev1.EventTypeWarning,
			"PausedReconciliation", fmt.Sprintf("label '%s' is set to true", pauseReconciliationLabel))

		patch := client.MergeFrom(visaInstance.DeepCopy())
		if writerErr := r.Status().Patch(ctx, visaInstance, patch); writerErr != nil {
			logger.Error(writerErr, "Error trying to Patch NoWarnings condition state")
		}
		return ctrl.Result{}, nil
	}

	if !visaInstance.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, visaInstance)
	}

	if err := r.ensureFinalizer(ctx, visaInstance); err != nil {
		return ctrl.Result{}, err
	}

	switch visaInstance.Spec.DesiredState {
	case cloudproviderv1.DesiredStateActive:
		result, err = r.reconcileActive(ctx, visaInstance)

	case cloudproviderv1.DesiredStateStopped:
		result, err = r.reconcileStopped(ctx, visaInstance)

	default:
		result = ctrl.Result{}
	}

	if err != nil {
		return result, err
	}

	return result, nil

}

func (r *VISAInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudproviderv1.VISAInstance{}).
		Named("visainstance").
		Owns(&corev1.Pod{}).
		Complete(r)
}

func (r *VISAInstanceReconciler) getVISAInstance(ctx context.Context, namespacedName types.NamespacedName) (*cloudproviderv1.VISAInstance, error) {
	visaInstance := &cloudproviderv1.VISAInstance{}
	err := r.Get(ctx, namespacedName, visaInstance)
	return visaInstance, err
}
