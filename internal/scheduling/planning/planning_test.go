package planning

import (
	"errors"
	"fmt"
	"testing"

	"github.com/SMALL-head/podGroup/internal/scheduling/model"
	"github.com/stretchr/testify/require"
)

func TestFunc(t *testing.T) {

	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			name: "TestPlanning",
			f:    case1,
		},
		{
			name: "TestRelativeImprovementAssign",
			f:    TestRelativeImprovementAssign,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, tc.f)
	}
}

func case1(t *testing.T) {
	// 9个pod
	podDependencies := new(model.PodDependencies)
	dependencyMatrix := [][]float64{
		{0, 1, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 1, 0, 0, 1, 0, 0, 1},
		{0, 0, 0, 1, 1, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 1, 1, 1},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	err := symmetryCopy(dependencyMatrix)
	require.NoError(t, err)
	podDependencies.BuildFromMatrix(dependencyMatrix)

	// 5个node
	nodeLatencies := new(model.NodeLatencies)
	nodeLatenciesMatrix := [][]float64{
		{0, 132, 121, 400, 130},
		{101, 0, 121, 400, 130},
		{101, 132, 0, 400, 130},
		{101, 132, 121, 0, 130},
		{417, 432, 321, 301, 0},
	}
	nodeLatencies.BuildFromMatrix(nodeLatenciesMatrix)

	nodes := []model.Node{
		{NodeName: "node1", CPUCap: 32, MemCap: 64},
		{NodeName: "node2", CPUCap: 32, MemCap: 64},
		{NodeName: "node3", CPUCap: 32, MemCap: 64},
		{NodeName: "node4", CPUCap: 32, MemCap: 64},
		{NodeName: "node5", CPUCap: 32, MemCap: 64},
	}

	pods := []model.PodModel{
		{PodName: "pod1", CPUReq: 2, MemReq: 4},
		{PodName: "pod2", CPUReq: 2, MemReq: 4},
		{PodName: "pod3", CPUReq: 2, MemReq: 4},
		{PodName: "pod4", CPUReq: 2, MemReq: 4},
		{PodName: "pod5", CPUReq: 2, MemReq: 4},
		{PodName: "pod6", CPUReq: 2, MemReq: 4},
		{PodName: "pod7", CPUReq: 2, MemReq: 4},
		{PodName: "pod8", CPUReq: 2, MemReq: 4},
		{PodName: "pod9", CPUReq: 2, MemReq: 4},
	}

	assign, score := SimulatedAnnealingAssign(0.3,
		0.7,
		*nodeLatencies,
		*podDependencies,
		pods, nodes, 100000, 200, 1, 0.95, true)
	fmt.Println("assign: ", assign)
	fmt.Println("score: ", score)

}

func TestRelativeImprovementAssign(t *testing.T) {
	// 9个pod
	podDependencies := new(model.PodDependencies)
	dependencyMatrix := [][]float64{
		{0, 1, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 1, 0, 0, 1, 0, 0, 1},
		{0, 0, 0, 1, 1, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 1, 1, 1},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	err := symmetryCopy(dependencyMatrix)
	require.NoError(t, err)
	podDependencies.BuildFromMatrix(dependencyMatrix)

	// 5个node
	nodeLatencies := new(model.NodeLatencies)
	nodeLatenciesMatrix := [][]float64{
		{0, 132, 121, 400, 130},
		{101, 0, 121, 400, 130},
		{101, 132, 0, 400, 130},
		{101, 132, 121, 0, 130},
		{417, 432, 321, 301, 0},
	}
	nodeLatencies.BuildFromMatrix(nodeLatenciesMatrix)

	nodes := []model.Node{
		{NodeName: "node1", CPUCap: 32, MemCap: 64},
		{NodeName: "node2", CPUCap: 32, MemCap: 64},
		{NodeName: "node3", CPUCap: 32, MemCap: 64},
		{NodeName: "node4", CPUCap: 32, MemCap: 64},
		{NodeName: "node5", CPUCap: 32, MemCap: 64},
	}

	pods := []model.PodModel{
		{PodName: "pod1", CPUReq: 2, MemReq: 4},
		{PodName: "pod2", CPUReq: 2, MemReq: 4},
		{PodName: "pod3", CPUReq: 2, MemReq: 4},
		{PodName: "pod4", CPUReq: 2, MemReq: 4},
		{PodName: "pod5", CPUReq: 2, MemReq: 4},
		{PodName: "pod6", CPUReq: 2, MemReq: 4},
		{PodName: "pod7", CPUReq: 2, MemReq: 4},
		{PodName: "pod8", CPUReq: 2, MemReq: 4},
		{PodName: "pod9", CPUReq: 2, MemReq: 4},
	}

	assign, score := RelativeImprovementAssign(0.7,
		0.3,
		*nodeLatencies,
		*podDependencies,
		pods, nodes, 10000, 1000, 0.1, 0.98, true)
	fmt.Println("assign: ", assign)
	fmt.Println("relative improvement: ", score)
}

func symmetryCopy(matrix [][]float64) error {
	if len(matrix) == 0 {
		return nil
	}
	m, n := len(matrix), len(matrix[0])
	if m != n {
		return errors.New("matrix is not square")
	}

	for i := 0; i < m; i++ {
		for j := i; j < n; j++ {
			matrix[j][i] = matrix[i][j]
		}
	}
	return nil
}
