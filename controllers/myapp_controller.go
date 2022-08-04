/*
Copyright 2022 cnych.

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
	appv1beta1 "github.com/cnych/opdemo/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	oldSpecAnnaAnnotation = "old/spec"
	log                   = logf.Log.WithName("myapp")
)

// MyAppReconciler reconciles a MyApp object
type MyAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.ydsz.io,resources=myapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.ydsz.io,resources=myapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.ydsz.io,resources=myapps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MyApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *MyAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// TODO(user): your logic here

	// 首先获取myapp 实例
	var myapp appv1beta1.MyApp
	if err := r.Client.Get(ctx, req.NamespacedName, &myapp); err != nil {
		// MyApp was deleted. Ignore
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 得到MyApp过后去创建对应的Deployment和Service
	// 创建就得去判断是否存在：存在则忽略，否则就创建
	// 更新：和上面一样

	// 调谐，获取当前的一个状态，然后和我们期望的状态对比
	// CreateOrUpdate Deployment
	var deploy appsv1.Deployment
	deploy.Name = myapp.Name
	deploy.Namespace = myapp.Namespace
	OR, err := ctrl.CreateOrUpdate(ctx, r.Client, &deploy, func() error {
		// 调谐必须在这个函数中去实现
		MutateDeployment(&myapp, &deploy)
		return controllerutil.SetControllerReference(&myapp, &deploy, r.Scheme)
	})

	if err != nil {
		return ctrl.Result{}, err
	}
	log.Info("CreateOrUpdate", "Deployment", OR)

	// CreateOrUpdate Service
	var svc corev1.Service
	svc.Name = myapp.Name
	svc.Namespace = myapp.Namespace
	OR, err = ctrl.CreateOrUpdate(ctx, r.Client, &svc, func() error {
		// 调谐必须在这个函数中去实现
		MutateService(&myapp, &svc)
		return controllerutil.SetControllerReference(&myapp, &svc, r.Scheme)
	})

	if err != nil {
		return ctrl.Result{}, err
	}
	log.Info("CreateOrUpdate", "Service", OR)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1beta1.MyApp{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
