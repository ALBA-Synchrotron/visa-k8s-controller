package controller

import (
	"context"

	cloudproviderv1 "git.cells.es/mis/visa/visa-k8s-controller/api/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *VISAInstanceReconciler) reconcilePod(ctx context.Context, desired *corev1.Pod) error {

	current := &corev1.Pod{}

	err := r.Get(ctx, client.ObjectKey{
		Namespace: desired.Namespace,
		Name:      desired.Name,
	}, current)

	switch {
	case apierrors.IsNotFound(err):
		return r.Create(ctx, desired)

	case err != nil:
		return err
	}

	desiredHash := desired.Annotations["cloudprovider.visa.k8s/spec-hash"]
	currentHash := current.Annotations["cloudprovider.visa.k8s/spec-hash"]

	if currentHash != desiredHash {
		return r.Delete(ctx, current)
	}

	return nil
}

func (r *VISAInstanceReconciler) deletePod(ctx context.Context, instance *cloudproviderv1.VISAInstance) error {
	pod := &corev1.Pod{}

	err := r.Get(ctx, client.ObjectKey{
		Namespace: instance.Namespace,
		Name:      instance.Name + "-runtime",
	}, pod)

	switch {
	case apierrors.IsNotFound(err):
		return nil

	case err != nil:
		return err
	}

	return r.Delete(ctx, pod)
}

func (r *VISAInstanceReconciler) getPod(ctx context.Context, instance *cloudproviderv1.VISAInstance) (*corev1.Pod, error) {

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
