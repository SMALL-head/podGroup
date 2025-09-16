/*
Copyright 2025.

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
	"slices"
	"time"

	"github.com/SMALL-head/podGroup/internal/prome"
	"github.com/SMALL-head/podGroup/internal/scheduling/model"
	"github.com/SMALL-head/podGroup/internal/scheduling/planning"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "github.com/SMALL-head/podGroup/api/v1"
)

// PodGroupReconciler reconciles a PodGroup object
type PodGroupReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	PromeClient *prome.PromClient
}

// +kubebuilder:rbac:groups=core.cic.io,resources=podgroups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.cic.io,resources=podgroups/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.cic.io,resources=podgroups/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=nodes;pods;services,verbs=get;list;watch;delete;create
// +kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="batch",resources=jobs,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PodGroup object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *PodGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	podGroup := &corev1.PodGroup{}
	err := r.Get(ctx, req.NamespacedName, podGroup)

	if err != nil {
		klog.Errorf("Failed to get PodGroup %s/%s, err: %v", req.Namespace, req.Name, err)
	}

	// 1. 解析dependencies
	pRes := planning.ParsePodGroup(podGroup)
	if pRes == nil {
		return ctrl.Result{}, nil
	}

	// 2. Pod按度数从高到低排序
	podNameListByDegree := planning.SortPodNameListByDegree(pRes.PodDependencies, pRes.PodNameList)

	// 3. 从可用节点中选择k的节点，保证最近延迟最小的k个节点
	// 3.1 获取最近5分钟的节点延迟数据
	end := time.Now()
	start := end.Add(-5 * time.Minute)
	resMatrix, err := r.PromeClient.GetLatencyByTimeRange(start.Format(time.RFC3339), end.Format(time.RFC3339))
	if err != nil {
		klog.Errorf("Failed to get latency from Prometheus, err: %v", err)
		// 降级为普通的调度模式
		_ = planning.NormalSchedule(ctx, r.Client, podGroup)
		return ctrl.Result{}, err
	}
	// klog.Infof("PodDependencies: %v", pRes.PodDependencies)
	// klog.Infof("PodNameList: %v", pRes.PodNameList)

	// klog.Infof("PodNameListByDegree: %v", podNameListByDegree)

	// 3.2 node延迟排序
	nodeLatencies := model.PrometheusMatrix2NodeLatencies(resMatrix)
	nodeNameList := make([]string, 0, len(nodeLatencies))
	for k := range nodeLatencies {
		nodeNameList = append(nodeNameList, k)
	}

	slices.SortFunc(nodeNameList, func(i, j string) int {
		if nodeLatencies[i] < nodeLatencies[j] {
			return -1
		} else if nodeLatencies[i] > nodeLatencies[j] {
			return 1
		}
		return 0
	})

	// klog.Infof("NodeNameList: %v", nodeNameList)

	// 4. 贪心placement
	podNodeMapper := planning.GreedyPlacement(podNameListByDegree, nodeNameList, pRes.NodeBalanceFactor)

	//打印podPerNode
	// klog.Infof("NodeBalanceFactor: %v", pRes.NodeBalanceFactor)

	// 5. placement采用nodeAffinity策略绑定节点，调度器在其他条件不符合的情况(例如cpu，mem资源不够)下调度至其他节点
	gvk, _, err := r.Scheme.ObjectKinds(podGroup)
	// 注： 这里的gvk是一个长度为1的数组，其中Group = "core.cic.io", Version = "v1", Kind = "PodGroup"
	if err != nil || len(gvk) == 0 {
		klog.Errorf("Failed to get GVK from Scheme, err: %v", err)
		return ctrl.Result{}, err
	}
	for podName := range podNodeMapper {
		podTemplate := pRes.PodGroupMap[podName]
		affinityNode := podNodeMapper[podName]
		pod := model.PodTemplate2PodSpec(podTemplate, podGroup.ObjectMeta, affinityNode, gvk[0])

		// 创建Pod
		klog.Infof("Creating Pod %s/%s on Node (Affinity) %s", pod.Namespace, pod.Name, affinityNode)
		err = r.Create(ctx, &pod)
		if err != nil {
			klog.Errorf("Failed to create Pod %s/%s, err: %v", pod.Namespace, pod.Name, err)
			return ctrl.Result{}, err
		}
	}

	// klog.Infof("[Reconcile] - get PodGroup %s/%s", req.Namespace, req.Name)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.PodGroup{}).
		Named("podgroup").
		Complete(r)
}
