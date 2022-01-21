/*
Copyright 2022.

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

package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// DeploymentReconciler reconciles a Deployment object
type DeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Deployment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
//
// TODO: ideally, the controller would watch on a custom resource (which will have the config)
func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var deployment appsv1.Deployment
	if err := r.Get(ctx, req.NamespacedName, &deployment); err != nil {
		if apierrors.IsNotFound(err) {
			// we'll ignore not-found errors, since we can get them on deleted requests.
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch Deployment")
		return ctrl.Result{}, err
	}

	// get hpa associated with this deployment (same name)
	// get current min max replicas from hpa
	// we have current replica from the deployment
	// final value would be the goal replicas to reach

	// calculate desired replica count (single value for both min/max)
	// dampen the value according to the step size as part of config

	// if desired_value < current_max; desired_value = current_max

	// if final_value == desired_value: return

	// else: update HPA resource with the new min max values

	// at any point, we're not going to make changes to the
	// the deployment resource, that will be taken care of
	// by the HPA resource

	// NOTE: this is going to cause a infinite loop
	// as it is right now, since the HPA will update the
	// deployment and the deployment will trigger this controller

	// IMP: we can check the available vs desired count in the deployment
	// if that is not equal, we can return and let it come up, once the
	// last pod has come up, we'll run the entire reconciliation logic
	// again

	// update deployment conditionally

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
// Tells the operator framework that we want to watch
// the deployment resource
func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Complete(r)
}
