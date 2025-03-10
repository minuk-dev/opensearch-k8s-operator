package controllers

import (
	"context"
	opensearchv1 "github.com/opensearch-project/opensearch-k8s-operator/opensearch-operator/api/v1"

	"github.com/go-logr/logr"
	"github.com/opensearch-project/opensearch-k8s-operator/opensearch-operator/pkg/reconcilers"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// OpensearchTenantReconciler reconciles a OpensearchTenant object
type OpensearchTenantReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Instance *opensearchv1.OpensearchTenant
	logr.Logger
}

//+kubebuilder:rbac:groups=opensearch.org,resources=opensearchtenants,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=opensearch.org,resources=opensearchtenants/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=opensearch.org,resources=opensearchtenants/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *OpensearchTenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Logger = log.FromContext(ctx).WithValues("tenant", req.NamespacedName)
	r.Logger.Info("Reconciling OpensearchTenant")

	r.Instance = &opensearchv1.OpensearchTenant{}
	err := r.Get(ctx, req.NamespacedName, r.Instance)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	tenantReconciler := reconcilers.NewTenantReconciler(
		r.Client,
		ctx,
		r.Recorder,
		r.Instance,
	)

	if r.Instance.DeletionTimestamp.IsZero() {
		controllerutil.AddFinalizer(r.Instance, OpensearchFinalizer)
		err = r.Client.Update(ctx, r.Instance)
		if err != nil {
			return ctrl.Result{}, err
		}
		return tenantReconciler.Reconcile()
	} else {
		if controllerutil.ContainsFinalizer(r.Instance, OpensearchFinalizer) {
			err = tenantReconciler.Delete()
			if err != nil {
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(r.Instance, OpensearchFinalizer)
			return ctrl.Result{}, r.Client.Update(ctx, r.Instance)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OpensearchTenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opensearchv1.OpensearchTenant{}).
		Owns(&opensearchv1.OpenSearchCluster{}). // Get notified when opensearch clusters change
		Complete(r)
}
