/*
Copyright 2023.

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
        "sort"
        "time"

        "github.com/robfig/cron"
        kbatch "k8s.io/api/batch/v1"
        corev1 "k8s.io/api/core/v1"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        "k8s.io/apimachinery/pkg/runtime"
        ref "k8s.io/client-go/tools/reference"
        ctrl "sigs.k8s.io/controller-runtime"
        "sigs.k8s.io/controller-runtime/pkg/client"
        "sigs.k8s.io/controller-runtime/pkg/log"

	maintenancev1alpha1 "nightly-worker/api/v1alpha1"
)

// NodeMaintenanceReconciler reconciles a NodeMaintenance object
type NodeMaintenanceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Clock
}

type realClock struct{}

func (_ realClock) Now() time.Time { return time.Now() }

// clock knows how to get the current time.
// It can be used to fake out timing for testing.
type Clock interface {
    Now() time.Time
}

//+kubebuilder:rbac:groups=maintenance.nightlyworker.com,resources=nodemaintenances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=maintenance.nightlyworker.com,resources=nodemaintenances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=maintenance.nightlyworker.com,resources=nodemaintenances/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get

var (
    scheduledTimeAnnotation = "nodemaintenances.maintenance.nightlyworker.com/scheduled-at"
)

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NodeMaintenance object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *NodeMaintenanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Load CronJob by name
	var cronJob batchv1.CronJob
        if err := r.Get(ctx, req.NamespacedName, &cronJob); err != nil {
                log.Error(err, "unable to fetch CronJob")
                // we'll ignore not-found errors, since they can't be fixed by an immediate
                // requeue (we'll need to wait for a new notification), and we can get them
                // on deleted requests.
                return ctrl.Result{}, client.IgnoreNotFound(err)
        }

	// List all active jobs:
	var childJobs kbatch.JobList
        if err := r.List(ctx, &childJobs, client.InNamespace(req.Namespace), client.MatchingFields{jobOwnerKey: req.Name}); err != nil {
            log.Error(err, "unable to list child Jobs")
            return ctrl.Result{}, err
        }



	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeMaintenanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&maintenancev1alpha1.NodeMaintenance{}).
		Complete(r)
}
