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
	"fmt"
	alpha1 "github.com/lvivJavaClub/kubebuilder-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// BlogAppReconciler reconciles a BlogApp object
type BlogAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=javaclub.lviv.ua,resources=blogapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=javaclub.lviv.ua,resources=blogapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=javaclub.lviv.ua,resources=blogapps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the BlogApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *BlogAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	b := &alpha1.BlogApp{}
	err := r.Client.Get(ctx, req.NamespacedName, b)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, nil
	}

	if !b.GetDeletionTimestamp().IsZero() {
		return ctrl.Result{}, nil
	}

	err = r.reconcilePvc(ctx, b)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.reconcileDeployment(ctx, b)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.reconcileService(ctx, b)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *BlogAppReconciler) reconcilePvc(ctx context.Context, b *alpha1.BlogApp) error {
	meta := metadata(b)
	meta.Name = meta.Name + "-pvc"
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: meta,
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: *b.Spec.Size,
				},
			},
		},
	}

	err := controllerutil.SetControllerReference(b, pvc, r.Scheme)
	if err != nil {
		return fmt.Errorf("failed to set owner reference on Service: %w", err)
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, pvc, func() error {
		return nil
	})
	return err
}

func (r *BlogAppReconciler) reconcileDeployment(ctx context.Context, b *alpha1.BlogApp) error {
	ls := labels(b)
	dep := &appsv1.Deployment{
		ObjectMeta: metadata(b),
		Spec: appsv1.DeploymentSpec{
			// Management Center StatefulSet size is always 1
			Replicas: &[]int32{1}[0],
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{{
						Name: b.Name + "-volume",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: b.Name + "-pvc",
								ReadOnly:  false,
							},
						},
					}},
					Containers: []corev1.Container{{
						Name: "blog-app",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 3001,
							Name:          "blog",
							Protocol:      corev1.ProtocolTCP,
						}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      b.Name + "-volume",
							MountPath: "/usr/blog/content",
						}},
						SecurityContext: &corev1.SecurityContext{
							RunAsNonRoot:             &[]bool{true}[0],
							RunAsUser:                &[]int64{65534}[0],
							Privileged:               &[]bool{false}[0],
							ReadOnlyRootFilesystem:   &[]bool{false}[0],
							AllowPrivilegeEscalation: &[]bool{false}[0],
							Capabilities: &corev1.Capabilities{
								Drop: []corev1.Capability{"ALL"},
							},
						},
					}},
					SecurityContext: &corev1.PodSecurityContext{
						FSGroup: &[]int64{65534}[0],
					},
				},
			},
		},
	}
	err := controllerutil.SetControllerReference(b, dep, r.Scheme)
	if err != nil {
		return fmt.Errorf("failed to set owner reference on StatefulSet: %w", err)
	}
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, dep, func() error {
		dep.Spec.Template.Spec.Containers[0].Image = "sbishyr/blog-backend:1.0"
		dep.Spec.Template.Spec.Containers[0].Env = []corev1.EnvVar{{Name: "CONTENT_DIR", Value: "/usr/blog/content/"}}
		dep.Spec.Template.Spec.Containers[0].ImagePullPolicy = corev1.PullAlways
		return nil
	})
	if errors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func (r *BlogAppReconciler) reconcileService(ctx context.Context, b *alpha1.BlogApp) error {
	svc := &corev1.Service{
		ObjectMeta: metadata(b),
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Port:       3001,
				TargetPort: intstr.FromInt(3001),
				Name:       "blog",
				Protocol:   corev1.ProtocolTCP,
			}},
			Selector: labels(b),
			Type:     b.Spec.ServiceType,
		},
	}

	err := controllerutil.SetControllerReference(b, svc, r.Scheme)
	if err != nil {
		return fmt.Errorf("failed to set owner reference on Service: %w", err)
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		return nil
	})
	return err
}

func labels(b *alpha1.BlogApp) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "lviv-java-club",
		"app.kubernetes.io/instance":   b.Name,
		"app.kubernetes.io/managed-by": "lvivjavaclub-operator",
	}
}

func metadata(b *alpha1.BlogApp) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      b.Name,
		Namespace: b.Namespace,
		Labels:    labels(b),
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *BlogAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&alpha1.BlogApp{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
