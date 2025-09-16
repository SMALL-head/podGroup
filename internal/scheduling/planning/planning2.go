package planning

import (
	"context"
	"sort"

	podGroupv1 "github.com/SMALL-head/podGroup/api/v1"
	"github.com/SMALL-head/podGroup/internal/scheduling/model"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/klog/v2"
)

// ParsePodGroup 解析PodGroup中的Pod及其依赖关系，返回PodGroupMap和PodDependencies
// PodGroupMap为一个map结构，key为pod的name，value为podTemplate
// PodDependencies为一个二维矩阵，表示Pod之间的通信关系
// Pod的名称列表，顺序与PodDependencies矩阵的行列顺序一致
// 若没有Pod，则返回nil
func ParsePodGroup(group *podGroupv1.PodGroup) *model.PodGroupParseResult {
	// TODO: 解析PodGroup中的Pod及其依赖关系，返回PodGroupMap和PodDependencies
	if group == nil || len(group.Spec.PodList) == 0 {
		return nil
	}
	// 构建PodGroupMap
	podGroupMap := make(model.PodGroupMap)
	for _, podTemplate := range group.Spec.PodList {
		podGroupMap[podTemplate.Metadata.Name] = podTemplate
	}
	// 构建PodNameList
	podNameList := make([]string, 0, len(podGroupMap))
	for podName := range podGroupMap {
		podNameList = append(podNameList, podName)
	}
	// 构建PodDependencies矩阵
	podCount := len(podNameList)
	podDependencies := make(model.PodDependencies, podCount)
	for i := range podDependencies {
		podDependencies[i] = make([]float64, podCount)
	}
	for _, dep := range group.Spec.Dependencies {
		var i, j int = -1, -1
		for idx, podName := range podNameList {
			if podName == dep.P1 {
				i = idx
			}
			if podName == dep.P2 {
				j = idx
			}
		}
		if i != -1 && j != -1 {
			podDependencies.Set(i, j, 1)
			podDependencies.Set(j, i, 1)
		}
	}
	//NodeBalanceFactor
	nodeBalanceFactor := group.Spec.NodeNum

	return &model.PodGroupParseResult{
		PodGroupMap:       podGroupMap,
		PodDependencies:   podDependencies,
		PodNameList:       podNameList,
		NodeBalanceFactor: nodeBalanceFactor,
	}
	//return nil
}

// SortPodNameListByDegree 根据Pod的依赖关系对Pod名称列表进行排序，返回排序后的Pod名称列表,按照度数从高到低排序
func SortPodNameListByDegree(dependencies model.PodDependencies, podNameList []string) []string {
	// TODO
	// 计算每个Pod的度数
	degrees := make([]int, len(podNameList))
	for i := range dependencies {
		for j := range dependencies[i] {
			if dependencies[i][j] == 1 {
				degrees[i]++
			}
		}
	}
	// 排序
	// sort.Slice(podNameList, func(i, j int) bool {
	// 	return degrees[i] > degrees[j]
	// })
	idx := make([]int, len(podNameList))
	for i := range idx {
		idx[i] = i
	}

	sort.Slice(idx, func(i, j int) bool {
		return degrees[idx[i]] > degrees[idx[j]]
	})

	sorted := make([]string, len(podNameList))
	for i, v := range idx {
		sorted[i] = podNameList[v]
	}
	return sorted
	// return nil
}

// NormalSchedule 对PodGroup进行常规调度 one-by-one
func NormalSchedule(ctx context.Context, c client.Client, group *podGroupv1.PodGroup) error {
	// TODO
	if group == nil || len(group.Spec.PodList) == 0 {
		return nil
	}
	// 获取PodGroup的GVK信息
	// if err != nil || len(gvk) == 0 {
	// }
	gvk := schema.GroupVersionKind{
		Group:   "core.cic.io",
		Version: "v1",
		Kind:    "PodGroup",
	}

	for _, podTemplate := range group.Spec.PodList {
		pod := model.CreatePodWithoutAffinity(podTemplate, group.ObjectMeta, gvk)

		klog.Infof("[NormalSchedule] Creating Pod %s/%s without NodeAffinity",
			pod.Namespace, pod.Name)

		if err := c.Create(ctx, &pod); err != nil {
			klog.Errorf("[NormalSchedule] Failed to create Pod %s/%s, err: %v",
				pod.Namespace, pod.Name, err)
			return err
		}
	}
	return nil
}

/*
GreedyPlacement 对PodGroup进行贪心调度，返回一个PodName - NodeName的映射
  - PodNameList - 按照度数从高到低排序的Pod名称列表
  - NodeNameList - 按照平均延迟从低到高排序的可用节点名称列表
  - PodPerNode - 每个节点上允许部署的Pod数量
*/
func GreedyPlacement(podNameList []string, nodeNameList []string, nodeBalance int) map[string]string {
	podPerNode := len(podNameList) / nodeBalance
	podPerNode1 := podPerNode + 1
	r := len(podNameList) % nodeBalance
	res := make(map[string]string)
	i, j, t := 0, 0, 0
	for i < len(podNameList) && j < len(nodeNameList) {
		res[podNameList[i]] = nodeNameList[j]
		t++
		i++
		if r > 0 {
			if t >= podPerNode1 {
				t = 0
				j++
				r--
			}
		} else {
			if t >= podPerNode {
				t = 0
				j++
			}
		}
	}
	return res
}
