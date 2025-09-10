package model

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

// PodDependencies 表示了Pod之间的通信关系
type PodDependencies = Matrix

// NodeLatencies 表示Node之间的通信延迟
type NodeLatencies = Matrix
