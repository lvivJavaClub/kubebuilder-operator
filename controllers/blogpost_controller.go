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
	alpha1 "github.com/lvivJavaClub/kubebuilder-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// BlogPostReconciler reconciles a BlogPost object
type BlogPostReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=javaclub.lviv.ua,resources=blogposts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=javaclub.lviv.ua,resources=blogposts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=javaclub.lviv.ua,resources=blogposts/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the BlogPost object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *BlogPostReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	bp := &alpha1.BlogPost{}
	err := r.Get(ctx, req.NamespacedName, bp)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if !bp.GetDeletionTimestamp().IsZero() {
		if err := deletePost(ctx, bp); err != nil {
			return ctrl.Result{}, err
		}
		controllerutil.RemoveFinalizer(bp, "lvivJavaClub.ua")
		err := r.Update(ctx, bp)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(bp, "lvivJavaClub.ua") {
		controllerutil.AddFinalizer(bp, "lvivJavaClub.ua")
		err := r.Update(ctx, bp)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	ba := &alpha1.BlogApp{}
	err = r.Get(ctx, types.NamespacedName{Name: bp.Spec.BlogAppName, Namespace: req.Namespace}, ba)
	if err != nil {
		return ctrl.Result{}, err
	}

	_, err = createNewPost(ctx, bp)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BlogPostReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&alpha1.BlogPost{}).
		Complete(r)
}
