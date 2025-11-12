package model

import (
	"strings"

	podGroupv1 "github.com/SMALL-head/podGroup/api/v1"
	"github.com/prometheus/common/model"
)

type Node struct {
	NodeName string
	CPUCap   float64
	MemCap   float64
}

type PodModel struct {
	PodName string
	CPUReq  float64
	MemReq  float64
}

type PodGroupMap map[string]podGroupv1.PodTemplate

// Matrix 通用二维浮点型矩阵
type Matrix [][]float64

func (m *Matrix) Set(i, j int, value float64) {
	// 若行未初始化，先分配空间
	if len(*m) <= i {
		return // 或者 panic("index out of range")
	}
	if len((*m)[i]) <= j {
		return // 或者 panic("index out of range")
	}
	(*m)[i][j] = value
}

func (m *Matrix) Get(i, j int) float64 {
	if len(*m) <= i || len((*m)[i]) <= j {
		return 0 // 或者 panic("index out of range")
	}
	return (*m)[i][j]
}

func (m *Matrix) BuildFromMatrix(matrix [][]float64) {
	*m = make([][]float64, len(matrix))
	for i := range matrix {
		(*m)[i] = make([]float64, len(matrix[i]))
		copy((*m)[i], matrix[i])
	}
}

type PodGroupParseResult struct {
	PodGroupMap       PodGroupMap
	PodDependencies   PodDependencies
	PodNameList       []string // 顺序与PodDependencies矩阵的行列顺序一致
	NodeBalanceFactor int
}

// PodDependencies 表示了Pod之间的通信关系
type PodDependencies = Matrix

// NodeLatencies 表示Node之间的通信延迟
type NodeLatencies = Matrix

type NodeTotalLatencies map[string]float64

func PrometheusMatrix2NodeLatencies(matrix model.Matrix) NodeTotalLatencies {
	res := make(NodeTotalLatencies)
	srcLen := make(map[string]int)
	for _, sample := range matrix {
		// 节点名称
		src := string(sample.Metric["src"])
		dst := string(sample.Metric["dst"])
		// 我们不计算control-plane节点的延迟，因为它通常不参与实际的工作负载通信
		if strings.Contains(src, "control") || strings.Contains(dst, "control") {
			continue
		}
		for _, v := range sample.Values {
			res[src] += float64(v.Value)
			res[dst] += float64(v.Value)
		}
		srcLen[src] += len(sample.Values)
		srcLen[dst] += len(sample.Values)
	}
	for k, v := range res {
		res[k] = v / float64(srcLen[k])
	}
	return res
}
