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

// STEP_SIZE denotes the step
// size for every subsequent scale
// operation on a deployment
const STEP_SIZE = 2

// Config temporarily stores
// goalReplicas for the specified
// runtimes.
type Config struct {
	Replicas int
}

// Ideally, we'd delegate all this
// configuration to the Custom Resource
var config = map[string]Config{
	"nodeservice": Config{
		Replicas: 10,
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

	// fetch the deployment manifest for
	// which the controller was invoked
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, req.NamespacedName, deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		if errors.IsResourceExpired(err) {
			log.Info("conflict, re-queuing again")
		}
		return reconcile.Result{}, err
	}

	// for testing, we're skipping other deployments
	if deployment.Name != "nodeservice" {
		return ctrl.Result{}, nil
	}

	goalReplicas := config[deployment.Name].Replicas
	currentReplicas := deployment.Status.Replicas
	updatedReplicas := deployment.Status.UpdatedReplicas
	availableReplicas := deployment.Status.AvailableReplicas

	// Checking two important conditions to skip the
	// reconciliation process:
	//   1. If the deployment has been updated previously
	//      by the controller and is still waiting on pods
	// 		to come up (updatedReplicas != availableReplicas)
	//		If we were to skip this, our controller might
	//		end up in an infinite loop trying to update
	// 		the deployment and the subsequent updates would
	//		trigger another reconciliation loop.
	//   2. If the deployment has reached the final goal
	//		replica after repeated reconciliations
	//		i.e. availableReplicas == goalReplicas
	//
	// We're checking in two separate conditionals just
	// to have the controller log the appropriate action
	if updatedReplicas != availableReplicas {
		log.Info("Deployment is waiting for pods to come up, skipping updates.", "deployment", deployment.Name)
		return ctrl.Result{}, nil
	} else if availableReplicas == int32(goalReplicas) {
		log.Info("Deployment up to date, skipping further updates.", "deployment", deployment.Name)
		return ctrl.Result{}, nil
	}

	// Fetch the HPA manifest corresponding to this deployment
	// Here, we're assuming they have the same name so we're not
	// referencing the deployment resource directly.
	hpa, err := r.getHPA(ctx, deployment.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.
			return reconcile.Result{}, nil
		}

		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	currMinReplicas := hpa.Spec.MinReplicas
	currMaxReplicas := hpa.Spec.MaxReplicas

	log.Info("Fetched HPA, calculating desiredReplicas", "HPA", hpa.Name)

	// Calculate the desiredReplicas for this deployment
	// based on the STEP_SIZE configuration option ideally
	// part of a custom resource. Since for our use-case,
	// we tend to set a single value for both min/max,
	// hence the single value for computation.
	//
	// Also worth exploring is dampening of the desiredReplicas
	// count if the delta is beyond a certain threshold (TODO)
	desiredReplicas := int(*currMinReplicas) * STEP_SIZE

	// If the desiredReplicas is lower than the
	// currentReplicas, make sure we don't reduce
	// the number of running replicas by lowering
	// the new max replica count. This could
	// end up having adverse impact on the application.
	//
	// In this case, it's okay to increase the min
	// count to the current replica size.
	if desiredReplicas < int(currentReplicas) {
		desiredReplicas = int(currentReplicas)
	}

	// If the desiredReplicas are lower than
	// the current max replicas, there should
	// not be a problem reaching this value.
	// This should ideally happen for when the scaling
	// has just started and we're running on our
	// normal HPA numbers.
	//
	// For e.g. let's say a workload has min/max 10/100
	// normally, and the controller comes up with a
	// value of 50 as the first step, then we can
	// match the max value here since we're already
	// reaching this max (100) as part of our usual
	// HPA operations.
	if desiredReplicas < int(currMaxReplicas) {
		desiredReplicas = int(currMaxReplicas)
	}

	// This is to ensure capping of the controller
	// operations essentially by checking if
	// the desiredReplicas comes out to be higher
	// than our goalReplicas. Just cap it to goalReplicas.
	if desiredReplicas > config[deployment.Name].Replicas {
		desiredReplicas = config[deployment.Name].Replicas
	}

	// Fetch HPA manifest again before updating it
	// so that we get the most up to date copy of it.
	// Changes that the controller makes to the
	// deployment controller would also trigger
	// HPA resource updates, so in order to avoid
	// conflicts during updates with respect to
	// validation (optimistic locking) from the API
	// server, be prudent and fetch the manifest again.
	hpa, err = r.getHPA(ctx, deployment.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Update the HPA resource with the
	// desired min/max values (desiredReplicas)
	t := int32(desiredReplicas)
	hpa.Spec.MinReplicas = &t
	hpa.Spec.MaxReplicas = int32(desiredReplicas)

	// Update the HPA for the current deployment
	log.Info("Updating HPA", "HPA", deployment.Name, "min/max", desiredReplicas)
	err = r.Update(ctx, hpa)
	if err != nil && !errors.IsAlreadyExists(err) {
		if errors.IsNotFound(err) {
			log.Info("HPA not found, requeuing request")
			return reconcile.Result{}, nil
		}

		if errors.IsConflict(err) {
			log.Info("HPA resourceVersion didn't match, requeuing request")
			return reconcile.Result{}, nil
		}

		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *DeploymentReconciler) getHPA(ctx context.Context, name string) (*autoscalingv1.HorizontalPodAutoscaler, error) {
	namespacesName := types.NamespacedName{
		Namespace: "default",
		Name:      name,
	}

	hpa := &autoscalingv1.HorizontalPodAutoscaler{}
	err := r.Get(ctx, namespacesName, hpa)
	if err != nil {
		return nil, err
	}
	return hpa, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Complete(r)
}
