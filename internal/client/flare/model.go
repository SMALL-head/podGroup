package flare

type SchedulingRecordRequest struct {
	Name         string `json:"name" binding:"required"`
	Namespace    string `json:"namespace" binding:"required"`
	ScheduledRes string `json:"schedule_res" binding:"required"`
	LatencyInfo  string `json:"latency_info" binding:"required"`
	UpdateAt     string `json:"update_at"`
	CommitTime   string `json:"commit_time" binding:"required"`
	UID          string `json:"uid" binding:"required"`
	Dependencies string `json:"dependencies"`
}

type SchedulingRecordStatusUpdateRequest struct {
	Name      string `json:"name" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	Status    string `json:"status" binding:"required"`
	UID       string `json:"uid"`
}

type LatencyMetric struct {
	Mi  float64 `json:"min"`
	Ma  float64 `json:"max"`
	Avg float64 `json:"avg"`
}
