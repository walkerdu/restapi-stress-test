package internal

import (
	"encoding/json"
	"sync"
)

type StressStat struct {
	TotalQueryCnts   int64
	TotalSuccessCnts int64
	TotalFailedCnts  int64

	QueryCnts      int64
	SuccessCnts    int64
	FailedCnts     int64
	FastestLatency int64
	SlowestLatency int64
	AvgLatency     int64
}

var mu sync.RWMutex

func (statIns *StressStat) Stat(success bool, latency int64) {
	mu.Lock()
	defer mu.Unlock()

	statIns.QueryCnts += 1
	statIns.TotalQueryCnts += 1

	if success {
		statIns.SuccessCnts += 1
		statIns.TotalSuccessCnts += 1
	} else {
		statIns.FailedCnts += 1
		statIns.TotalFailedCnts += 1
	}

	// only statistic latency data when success
	if success {
		if statIns.FastestLatency == 0 || latency < statIns.FastestLatency {
			statIns.FastestLatency = latency
		}
		if latency > statIns.SlowestLatency {
			statIns.SlowestLatency = latency
		}

		statIns.AvgLatency += latency
	}
}

func (statIns *StressStat) String() string {
	mu.Lock()
	defer mu.Unlock()

	if statIns.TotalQueryCnts > 0 && statIns.SuccessCnts > 0 {
		statIns.AvgLatency /= statIns.SuccessCnts
	}

	var print string
	//if body, err := json.MarshalIndent(statIns, "", "\t"); err != nil {
	if body, err := json.Marshal(statIns); err != nil {
		print = err.Error()
	} else {
		print = string(body)
	}

	statIns.QueryCnts = 0
	statIns.SuccessCnts = 0
	statIns.FailedCnts = 0
	statIns.FastestLatency = 0
	statIns.SlowestLatency = 0
	statIns.AvgLatency = 0

	return print
}
