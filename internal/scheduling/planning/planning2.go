package planning

import (
	podGroupv1 "github.com/SMALL-head/podGroup/api/v1"
	"github.com/SMALL-head/podGroup/internal/scheduling/model"
)

// ParsePodGroup 解析PodGroup中的Pod及其依赖关系，返回PodGroupMap和PodDependencies
// PodGroupMap为一个map结构，key为pod的name，value为podTemplate
// PodDependencies为一个二维矩阵，表示Pod之间的通信关系
// Pod的名称列表，顺序与PodDependencies矩阵的行列顺序一致
// 若没有Pod，则返回nil
func ParsePodGroup(group *podGroupv1.PodGroup) *model.PodGroupParseResult {
	// TODO: 解析PodGroup中的Pod及其依赖关系，返回PodGroupMap和PodDependencies
	return nil
}

// SortPodNameListByDegree 根据Pod的依赖关系对Pod名称列表进行排序，返回排序后的Pod名称列表,按照度数从高到低排序
func SortPodNameListByDegree(dependencies model.PodDependencies, podNameList []string) []string {
	// TODO
	return nil
}

// NormalSchedule 对PodGroup进行常规调度 one-by-one
func NormalSchedule(group *podGroupv1.PodGroup) error {
	// TODO
	return nil
}

/*
GreedyPlacement 对PodGroup进行贪心调度，返回一个PodName - NodeName的映射
  - PodNameList - 按照度数从高到低排序的Pod名称列表
  - NodeNameList - 按照平均延迟从低到高排序的可用节点名称列表
  - PodPerNode - 每个节点上允许部署的Pod数量
*/
func GreedyPlacement(podNameList []string, nodeNameList []string, podPerNode int) map[string]string {
	res := make(map[string]string)
	i, j, t := 0, 0, 0
	for i < len(podNameList) && j < len(nodeNameList) {
		res[podNameList[i]] = nodeNameList[j]
		t++
		i++
		if t >= podPerNode {
			t = 0
			j++
		}
	}
	return res
}
