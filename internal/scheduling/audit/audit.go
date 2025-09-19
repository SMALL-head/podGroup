package audit

import (
	"bytes"
	"context"

	podGroupv1 "github.com/SMALL-head/podGroup/api/v1"
	"github.com/SMALL-head/podGroup/internal/client/flare"
	"github.com/SMALL-head/podGroup/internal/client/prome"
	"k8s.io/apimachinery/pkg/util/json"
)

func ReportLatencyInfo(pc *prome.PromClient, flareC *flare.Client, start, end string, pg *podGroupv1.PodGroup) error {
	latencyStatus, err := pc.GetLatencyStats(start, end)
	if err != nil {
		return nil
	}
	req := &flare.SchedulingRecordRequest{
		Name:        pg.Name,
		Namespace:   pg.Namespace,
		LatencyInfo: latencyStatus,
		UID:         string(pg.UID),
	}
	r, err := json.Marshal(req)
	if err != nil {
		return err
	}
	reqReader := bytes.NewReader(r)
	httpReq, err := flareC.NewRequest(context.Background(), "POST", "/cluster/scheduling/updateRecord/8", reqReader)
	if err != nil {
		return err
	}

	if _, err = flareC.Do(httpReq); err != nil {
		return err
	}

	return nil
}
