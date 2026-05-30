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
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	appsv1alpha1 "github.com/dev/labelSyncOperator/api/v1alpha1"
)

// SyncOperatorReconciler reconciles a SyncOperator object
type SyncOperatorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.dev.com,resources=syncoperators,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.dev.com,resources=syncoperators/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.dev.com,resources=syncoperators/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SyncOperator object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.23.3/pkg/reconcile
func (r *SyncOperatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch your CR
	var syncOperator appsv1alpha1.SyncOperator

	if err := r.Get(ctx, req.NamespacedName, &syncOperator); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}

	log.Info("Reconciling SyncOperator", "Name", syncOperator.Name, "Namespace", syncOperator.Namespace)

	// Get the deployment specified in the SyncOperator spec
	var targetDeployment appsv1.Deployment

	if err := r.Get(ctx, types.NamespacedName{Name: syncOperator.Spec.TargetDeployment,
		Namespace: syncOperator.Namespace}, &targetDeployment); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Target deployment not found, requeuing", "Deployment", syncOperator.Spec.TargetDeployment)
			return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
		}

		log.Error(err, "Failed to get target deployment", "Deployment", syncOperator.Spec.TargetDeployment)
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}

	MatchLabels := targetDeployment.Spec.Selector.MatchLabels

	// Get the pod based on the deployment's label selector
	var podList corev1.PodList
	if err := r.List(ctx, &podList, client.InNamespace(syncOperator.Namespace), client.MatchingLabels(MatchLabels)); err != nil {
		log.Error(err, "Failed to list pods for target deployment", "Deployment", syncOperator.Spec.TargetDeployment)
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}

	// Check if the labels on the pods match the labels specified in the TargetLabels and update them if they don't
	StatusMismatch := false
	TargetLabels := targetDeployment.Spec.Template.Labels
	for _, pod := range podList.Items {
		original := pod.DeepCopy() // Create a deep copy of the original pod to use for patching
		mismatch := false

		if pod.Labels == nil {
			pod.Labels = make(map[string]string)
		}

		for key, value := range TargetLabels {
			if pod.Labels[key] != value {
				StatusMismatch = true
				mismatch = true
				pod.Labels[key] = value
			}
		}

		if mismatch {
			if err := r.Patch(ctx, &pod, client.MergeFrom(original)); err != nil {
				log.Error(err, "Failed to patch pod labels", "Pod", pod.Name)
				return ctrl.Result{RequeueAfter: 30 * time.Second}, err
			}

			log.Info("Patched pod labels to match target deployment", "Pod", pod.Name)
		}
	}

	// Update the Sync status in the SyncOperator CR based on whether there was a mismatch
	if syncOperator.Status.Sync != !StatusMismatch {
		syncOperator.Status.Sync = !StatusMismatch

		if err := r.Status().Update(ctx, &syncOperator); err != nil {
			log.Error(err, "Failed to update SyncOperator status")
			return ctrl.Result{RequeueAfter: 30 * time.Second}, err
		}
	}
	log.Info("Updated SyncOperator status", "Sync", syncOperator.Status.Sync)

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SyncOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.SyncOperator{}).
		Named("syncoperator").
		Complete(r)
}
