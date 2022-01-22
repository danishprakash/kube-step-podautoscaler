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
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	types "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// DeploymentReconciler reconciles a Deployment object
type DeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const STEP_SIZE = 3

type Config struct {
	Replicas int
}

var config = map[string]Config{
	"execute-d": Config{
		Replicas: 50,
	},
}

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Deployment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// we have current replica from the deployment
	// final value would be the goal replicas to reach
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, req.NamespacedName, deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// TODO: for testing, only operate upon this
	if deployment.Name != "execute-d" {
		return ctrl.Result{}, nil
	}
	currentReplicas := deployment.Status.Replicas
	updatedReplicas := deployment.Status.UpdatedReplicas
	availableReplicas := deployment.Status.AvailableReplicas

	// Previous scale up event still pending
	// NOTE: this is going to cause a infinite loop
	// as it is right now, since the HPA will update the
	// deployment and the deployment will trigger this controller

	// IMP: we can check the available vs desired count in the deployment
	// if that is not equal, we can return and let it come up, once the
	// last pod has come up, we'll run the entire reconciliation logic
	// again
	// update deployment conditionally
	if updatedReplicas != availableReplicas {
		// TODO: add logs here
		return ctrl.Result{}, nil
	}

	namespacesName := types.NamespacedName{
		Namespace: "default",
		Name:      deployment.Name,
	}

	// get hpa associated with this deployment (same name)
	// get current min max replicas from hpa
	hpa := &autoscalingv1.HorizontalPodAutoscaler{}
	err = r.Get(ctx, namespacesName, hpa)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	minReplicas := hpa.Spec.MinReplicas
	maxReplicas := hpa.Spec.MaxReplicas

	// calculate desired replica count (single value for both min/max)
	// dampen the value according to the step size as part of config
	desiredReplicas := int(*minReplicas) * STEP_SIZE

	// if desired_value < current_max; desired_value = current_max
	if desiredReplicas < int(currentReplicas) {
		desiredReplicas = int(currentReplicas)
	}

	newMinReplicas := desiredReplicas
	newMaxReplicas := currentReplicas
	if newMaxReplicas < maxReplicas {
		newMaxReplicas = maxReplicas
	} else {
		newMaxReplicas = int32(desiredReplicas)
	}

	if newMinReplicas > int(newMaxReplicas) {
		newMinReplicas = int(newMaxReplicas)
	}

	// TODO: IMP: upper limit isn't being respected right now
	// if final_value == desired_value: return
	// we've reached our goal values, don't do anything
	if config[deployment.Name].Replicas == desiredReplicas {
		return ctrl.Result{}, nil
	}

	hpa.Spec.MaxReplicas = newMaxReplicas
	tmp := int32(newMinReplicas)
	hpa.Spec.MinReplicas = &tmp

	// else: update HPA resource with the new min max values
	// TODO: handle reconciliation errrors here
	err = r.Update(ctx, hpa)
	if err != nil && !errors.IsAlreadyExists(err) {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			log.Info("HPA not found, requeuing request")
			return reconcile.Result{}, nil
		}

		log.Info("HPA resourceVersion didn't match, requeuing request")
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	log.Info("values", "newMinReplicas", newMinReplicas, "newMaxReplicas", newMaxReplicas)

	// at any point, we're not going to make changes to the
	// the deployment resource, that will be taken care of
	// by the HPA resource

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Complete(r)
}
