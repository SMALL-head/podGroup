package planning

import (
	"fmt"
	"math"
	"math/rand"
	"path/filepath"
	"slices"
	"time"

	"github.com/SMALL-head/podGroup/internal/scheduling/model"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

const (
	ResourceLimitConstraint float64 = 100000000
)

// computeTotalLatency计算当前给定状态的延迟分数
// assign[i]表示Pod i 被分配到的节点编号
// dependencies[i][j]表示Pod i 和 Pod j 之间的通信需求，可以假设该数组是一个对称矩阵，且对角线为0（pod依赖关系不存在自环）
func computeTotalLatency(assign []int, latencies model.NodeLatencies, dependencies model.PodDependencies, size int) (score float64) {
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			nodeFori := assign[i]
			nodeForj := assign[j]
			score += latencies[nodeFori][nodeForj] * dependencies[i][j]
		}
	}

	return score / 2
}

func computeResourceBalancePenalty(assign []int, podModel []model.PodModel, nodeStatus []model.Node) (score float64, ok bool) {
	usedCPU := make([]float64, len(nodeStatus))
	usedMem := make([]float64, len(nodeStatus))

	ok = true

	for i := range assign {
		nodeIdx := assign[i]
		usedCPU[nodeIdx] += podModel[i].CPUReq
		usedMem[nodeIdx] += podModel[i].MemReq
	}

	// 计算惩罚分数
	for i := range nodeStatus {
		cpuUsageRatio := usedCPU[i] / nodeStatus[i].CPUCap
		memUsageRatio := usedMem[i] / nodeStatus[i].MemCap

		if cpuUsageRatio > 1 || memUsageRatio > 1 {
			ok = false
		}

		// 惩罚函数：使用平方惩罚，防止过度使用某个节点
		score += cpuUsageRatio * cpuUsageRatio
		score += memUsageRatio * memUsageRatio
	}

	return
}

func computeAllocBalancePenalty(assign []int, podModel []model.PodModel, nodeStatus []model.Node, latencyMap model.NodeLatencies) (score float64) {
	nodeSize := len(nodeStatus)
	lw := make([]float64, nodeSize)
	resourceCondition := make([]struct {
		usedCPU float64
		usedMem float64
		CPUCap  float64
		MemCap  float64
	}, nodeSize)
	for p, n := range assign {
		resourceCondition[n].usedCPU += podModel[p].CPUReq
		resourceCondition[n].usedMem += podModel[p].MemReq
	}
	for i := 0; i < nodeSize; i++ {
		latSum := 0.0
		for j := 0; j < nodeSize; j++ {
			latSum += latencyMap[i][j]
		}
		lw[i] = 1 / (latSum)
		resourceCondition[i].CPUCap = nodeStatus[i].CPUCap
		resourceCondition[i].MemCap = nodeStatus[i].MemCap
	}

	miu := 0.0
	for i := 0; i < nodeSize; i++ {
		ratio := (resourceCondition[i].usedCPU/resourceCondition[i].CPUCap + resourceCondition[i].usedMem/resourceCondition[i].MemCap) / 2
		miu += ratio / lw[i]
	}
	miu /= float64(nodeSize)
	for i := 0; i < nodeSize; i++ {
		score += (1/lw[i] - miu) * (1/lw[i] - miu)
	}
	return
}

// objectiveFunc 计算目标方程的值
func objectiveFunc(alpha float64, beta float64,
	latenciesMap model.NodeLatencies, podDependencies model.PodDependencies,
	pods []model.PodModel, nodeStatuses []model.Node,
	assign []int, podSize int) float64 {
	l := computeTotalLatency(assign, latenciesMap, podDependencies, podSize)
	p, ok := computeResourceBalancePenalty(assign, pods, nodeStatuses)
	if !ok {
		return ResourceLimitConstraint
	}
	//return l * p
	return alpha*l + beta*p
}

func objectFunc2(alpha float64, beta float64,
	latenciesMap model.NodeLatencies, podDependencies model.PodDependencies,
	pods []model.PodModel, nodeStatuses []model.Node,
	assign []int, podSize int,
	latencyMin, latencyMax float64, imbalanceMin, imbalanceMax float64,
) (float64, int) {
	l := computeTotalLatency(assign, latenciesMap, podDependencies, podSize)
	_, ok := computeResourceBalancePenalty(assign, pods, nodeStatuses)
	if !ok {
		return ResourceLimitConstraint, 0
	}
	p2 := computeAllocBalancePenalty(assign, pods, nodeStatuses, latenciesMap)

	fmt.Printf("assign: %v, latency factor: %f, imbalance factor: %f\n", assign, l, p2)

	// min-max归一化
	// 避免除0
	if latencyMax == latencyMin {
		latencyMax += 1
	}
	if imbalanceMax == imbalanceMin {
		imbalanceMax += 1
	}
	lOne := (l - latencyMin) / (latencyMax - latencyMin)
	pOne := (p2 - imbalanceMin) / (imbalanceMax - imbalanceMin)
	fmt.Printf("assign: %v, lOne factor: %f, pOne: %f\n", assign, lOne, pOne)
	var mode int
	if lOne*alpha > pOne*beta {
		mode = 1 // 延迟占比更大
	} else {
		mode = 2 // 资源均衡占比更大
	}
	return lOne + pOne, mode
}

// FindOptimalAssign 暴力枚举所有可能的assign，返回目标函数最小的分配方案
func FindOptimalAssign(
	alpha float64, beta float64,
	latenciesMap model.NodeLatencies, podDependencies model.PodDependencies,
	pods []model.PodModel, nodeStatuses []model.Node,
) (bestAssign []int, bestScore float64) {
	podSize := len(pods)
	nodeSize := len(nodeStatuses)
	if podSize == 0 || nodeSize == 0 {
		return nil, 0
	}

	assign := make([]int, podSize)
	bestAssign = make([]int, podSize)
	bestScore = 1e100 // 足够大的初始值

	var dfs func(idx int)
	dfs = func(idx int) {
		if idx == podSize {
			curScore := objectiveFunc(0.4, 0.6, latenciesMap, podDependencies, pods, nodeStatuses, assign, podSize)
			bestScore = min(curScore, bestScore)
			return
		}

		for i, _ := range nodeStatuses {
			assign[idx] = i
			dfs(idx + 1)
		}

	}

	dfs(0)
	return bestAssign, bestScore
}

// SimulatedAnnealingAssign 使用模拟退火算法搜索局部最优assign
func SimulatedAnnealingAssign(
	alpha float64, beta float64,
	latenciesMap model.NodeLatencies, podDependencies model.PodDependencies,
	pods []model.PodModel, nodeStatuses []model.Node,
	maxIter int, initTemp, finalTemp, coolingRate float64,
	debugMode bool,
) (bestAssign []int, bestScore float64) {
	podSize := len(pods)
	nodeSize := len(nodeStatuses)
	if podSize == 0 || nodeSize == 0 {
		return nil, 0
	}

	rand.Seed(time.Now().UnixNano())
	// 初始化assign
	assign := make([]int, podSize)
	for i := range assign {
		assign[i] = rand.Intn(nodeSize)
	}
	bestAssign = make([]int, podSize)
	copy(bestAssign, assign)
	bestScore = objectiveFunc(alpha, beta, latenciesMap, podDependencies, pods, nodeStatuses, assign, podSize)
	currScore := bestScore

	temp := initTemp

	var scoreCurve []float64
	if debugMode {
		scoreCurve = append(scoreCurve, currScore)
	}

	for iter := 0; iter < maxIter && temp > finalTemp; iter++ {
		// 生成新解：随机选一个Pod，分配到另一个Node
		newAssign := make([]int, podSize)
		copy(newAssign, assign)
		podIdx := rand.Intn(podSize)
		newNode := rand.Intn(nodeSize)
		for newNode == assign[podIdx] && nodeSize > 1 {
			newNode = rand.Intn(nodeSize)
		}
		newAssign[podIdx] = newNode

		newScore := objectiveFunc(alpha, beta, latenciesMap, podDependencies, pods, nodeStatuses, newAssign, podSize)
		delta := newScore - currScore

		accept := false
		if delta < 0 {
			accept = true
		} else {
			prob := math.Exp(-delta / temp)
			if rand.Float64() < prob {
				accept = true
			}
		}

		if accept {
			copy(assign, newAssign)
			currScore = newScore
			if currScore < bestScore {
				bestScore = currScore
				copy(bestAssign, assign)
			}
		}

		if debugMode {
			scoreCurve = append(scoreCurve, currScore)
		}

		temp *= coolingRate
	}

	if debugMode && len(scoreCurve) > 1 {
		_ = plotScoreCurve(scoreCurve)
	}

	return bestAssign, bestScore
}

// RelativeImprovementAssign 使用baseline归一化目标函数，返回相对改进���优的分配方案
func RelativeImprovementAssign(
	alpha float64, beta float64,
	latenciesMap model.NodeLatencies, podDependencies model.PodDependencies,
	pods []model.PodModel, nodeStatuses []model.Node,
	maxIter int, initTemp, finalTemp, coolingRate float64,
	debugMode bool) (bestAssign []int, bestScore float64) {
	podSize := len(pods)
	nodeSize := len(nodeStatuses)
	if podSize == 0 || nodeSize == 0 {
		return nil, 0
	}
	laMin, laMax := computeLatencyMinMax(latenciesMap, podDependencies)
	pMin, pMax := computePenaltyMinMax(pods, nodeStatuses, latenciesMap)

	rand.Seed(time.Now().UnixNano())
	// 初始化assign
	assign := make([]int, podSize)
	for i := range assign {
		assign[i] = rand.Intn(nodeSize)
	}
	assign[3] = 4
	assign[4] = 4
	assign[6] = 4
	assign[7] = 4
	bestAssign = make([]int, podSize)
	copy(bestAssign, assign)
	bestScore, mode := objectFunc2(alpha, beta, latenciesMap, podDependencies, pods, nodeStatuses, assign, podSize, laMin, laMax, pMin, pMax)
	currScore := bestScore

	temp := initTemp

	var scoreCurve []float64
	repeatTime := 0

	for iter := 0; iter < maxIter && temp > finalTemp; iter++ {
		// 温度高于0.4 * initTemp时，完全随机搜索
		var newAssign []int
		if repeatTime > 3 {
			repeatTime = 0
			newAssign = heuristicAssign(assign, latenciesMap, podDependencies, pods, nodeStatuses, mode)
		} else if temp > 0.3*initTemp {
			newAssign = randomAssign(assign, nodeStatuses)
		} else {
			newAssign = heuristicAssign(assign, latenciesMap, podDependencies, pods, nodeStatuses, mode)
		}
		var newScore float64
		var m int
		if temp > 0.4*initTemp {
			newScore, m = objectFunc2(0.7, 0.3, latenciesMap, podDependencies, pods, nodeStatuses, newAssign, podSize, laMin, laMax, pMin, pMax)
		} else {
			newScore, m = objectFunc2(0.5, 0.6, latenciesMap, podDependencies, pods, nodeStatuses, newAssign, podSize, laMin, laMax, pMin, pMax)
		}
		mode = m
		delta := newScore - currScore

		accept := false
		if delta < 0 {
			accept = true
		} else {
			prob := math.Exp(-delta / temp)
			if rand.Float64() < prob {
				accept = true
			}
		}

		if accept {
			copy(assign, newAssign)
			currScore = newScore
			if currScore < bestScore {
				bestScore = currScore
				copy(bestAssign, assign)
			}
		} else {
			repeatTime++
		}

		if debugMode {
			scoreCurve = append(scoreCurve, currScore)
		}

		temp *= coolingRate
	}

	if debugMode && len(scoreCurve) > 1 {
		_ = plotScoreCurve(scoreCurve)
	}

	return bestAssign, bestScore
}

func computeLatencyMinMax(latenciesMap model.NodeLatencies, pods model.PodDependencies) (mi, max float64) {
	maxLatency := float64(0)
	for i := range latenciesMap {
		for _, each := range latenciesMap[i] {
			if each > maxLatency {
				maxLatency = each
			}
		}
	}

	// 统计pod中边之和
	edgeSum := float64(0)
	for i := range pods {
		for j := i + 1; j < len(pods); j++ {
			edgeSum += pods[i][j]
		}
	}

	return 0, edgeSum * maxLatency / 1.5
}

func computePenaltyMinMax(pods []model.PodModel, nodes []model.Node, latencyMap model.NodeLatencies) (mi, ma float64) {
	var cpuReqSum, memReqSum float64
	for _, p := range pods {
		cpuReqSum += p.MemReq
		memReqSum += p.MemReq
	}
	cpuRatio, memRatio := cpuReqSum/float64(len(nodes)), memReqSum/float64(len(nodes))
	nodeSize := len(nodes)
	lw := make([]float64, nodeSize)

	for i := 0; i < nodeSize; i++ {
		latSum := 0.0
		for j := 0; j < nodeSize; j++ {
			latSum += latencyMap[i][j]
		}
		lw[i] = 1 / (latSum)
	}
	// 最小值出现在pod均匀分配的情况下
	miu := float64(0)
	for i := 0; i < nodeSize; i++ {
		miu += (cpuRatio + memRatio) / 2 / lw[i]
	}
	miu /= float64(nodeSize)

	for i := 0; i < nodeSize; i++ {
		mi += ((cpuRatio+memRatio)/2/lw[i] - miu) * ((cpuRatio+memRatio)/2/lw[i] - miu)
	}

	// 最大值出现在所有pod分配在延迟最低的那个节点上
	lwMin := slices.Max(lw)
	miuMax := lwMin / float64(nodeSize)
	for i := 0; i < nodeSize; i++ {
		if lw[i] != lwMin {
			ma += miuMax * miuMax
		} else {
			ma += (1/lwMin - miuMax) * (1/lwMin - miuMax)
		}
	}
	return
}

// heuristicAssign 启发式分配算法，对一个originalAssign进行局部优化。当模拟退火算法降温到某个T时，从完全随机变为启发式局部优化。或者当模拟退火算法收敛后（连续x次迭代没有更优解），启发式局部优化
// mode = 0表示完全启发式；mode = 1表示上次延迟太高了，本次使用延迟降低为主的启发式； mode = 2表示上次资源不均衡，本次使用资源均衡为主的启发式
func heuristicAssign(originalAssign []int,
	latenciesMap model.NodeLatencies, podDependencies model.PodDependencies,
	pods []model.PodModel, nodeStatuses []model.Node,
	mode int) (newAssign []int) {
	// 随机一个pod，分配到其他node延迟更低的节点
	nodeSize := len(nodeStatuses)
	type lat struct {
		nodeIdx int
		latSum  float64
	}

	newAssign = make([]int, len(originalAssign))
	copy(newAssign, originalAssign)

	// 计算所有node 的邻域延迟
	nodeLatSumUp := make([]lat, nodeSize)

	for i, node := range latenciesMap {
		nodeLatSumUp[i].nodeIdx = i
		for j := 0; j < len(node); j++ {
			nodeLatSumUp[i].latSum += node[j]
		}
	}

	slices.SortFunc(nodeLatSumUp, func(a, b lat) int {
		if a.latSum-b.latSum >= 0 {
			return 1
		} else {
			return -1
		}
	})

	if mode == 1 { // 延迟较高，微调度延迟较低的节点
		podIdx := rand.Intn(len(newAssign))
		currentNode := newAssign[podIdx]
		for i, l := range nodeLatSumUp {
			if l.nodeIdx == currentNode {
				if i > 0 {
					newAssign[podIdx] = nodeLatSumUp[i-1].nodeIdx
					fmt.Printf("启发式assign, mode=1, pod %d 从节点 %d 迁移到节点 %d\n", podIdx, currentNode, newAssign[podIdx])
					break
				}
			}
		}
	} else if mode == 2 {
		podDeg := make([]struct {
			podIdx int
			deg    float64
		}, len(pods))
		for i, depi := range podDependencies {
			podDeg[i].podIdx = i
			for j := i + 1; j < len(depi); j++ {
				if depi[j] > 0 {
					podDeg[i].deg += depi[j]
					podDeg[j].deg += depi[j]
				}
			}
		}
		slices.SortFunc(podDeg, func(a, b struct {
			podIdx int
			deg    float64
		}) int {
			if a.deg-b.deg >= 0 {
				return 1
			} else {
				return -1
			}
		})

		// 选度数最低的pod，且该pod分配的节点有重复的，进行变更

		// 使用bucket记录每个节点上的pod数量
		bucket := make([]int, nodeSize)

		for _, n := range newAssign {
			bucket[n]++
		}

		balance := float64(len(pods)) / float64(nodeSize)

		for _, each := range podDeg[:len(podDeg)-2] {
			if float64(bucket[newAssign[each.podIdx]]) >= balance+1 {
				newAssign[each.podIdx] = nodeLatSumUp[rand.Intn(nodeSize)].nodeIdx
				fmt.Printf("启发式assign, mode=2, pod %d 从节点 %d 迁移到节点 %d\n", each.podIdx, newAssign[each.podIdx], newAssign[each.podIdx])
				break
			}
		}
	} else {
		newAssign = randomAssign(newAssign, nodeStatuses)
	}
	return
}

func randomAssign(originalAssign []int, nodeStatuses []model.Node) (newAssign []int) {
	newAssign = make([]int, len(originalAssign))
	copy(newAssign, originalAssign)
	nodeSize := len(nodeStatuses)
	podIdx := rand.Intn(len(originalAssign))
	newNode := rand.Intn(nodeSize)
	for newNode == newAssign[podIdx] && nodeSize > 1 {
		newNode = rand.Intn(nodeSize)
	}
	return
}

// plotScoreCurve 绘制score变化曲线并保存到临时目录
func plotScoreCurve(scores []float64) string {
	p := plot.New()
	p.Title.Text = "Score Curve"
	p.X.Label.Text = "Iteration"
	p.Y.Label.Text = "Score"

	pts := make(plotter.XYs, len(scores))
	for i, v := range scores {
		pts[i].X = float64(i)
		pts[i].Y = v
	}

	err := plotutil.AddLinePoints(p, "Score", pts)
	if err != nil {
		return ""
	}

	filePath := filepath.Join("/Users/zyc/GolandProjects/podGroup/tmp", "score_curve.png")
	_ = p.Save(8*vg.Inch, 4*vg.Inch, filePath)
	return filePath
}
