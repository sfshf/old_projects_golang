package statistic

import "time"

type OracleFieldType struct {
	Duration int64 `json:"duration"` // millisecond
}

type StatisticInfo struct {
	Timestamp       time.Time `json:"timestamp"`
	ResponseCode    int32     `json:"responseCode"`
	Duration        int64     `json:"duration"` // millisecond
	Application     string    `json:"application"`
	Service         string    `json:"service"`
	Path            string    `json:"path"`
	ServiceDuration int64     `json:"serviceDuration"` // millisecond
}

const (
	RedisKeyPrefixStatisticInfo = "statisticinfo::"
)

type ProtoStatisticHourly struct {
	Timestamp            int64  `json:"date"`
	Application          string `json:"application"`
	Service              string `json:"service"`
	Path                 string `json:"path"`
	Hit                  int64  `json:"hit"`
	SuccessHit           int64  `json:"successHit"`
	ProxySuccessHit      int64  `json:"proxySuccessHit"`
	DurationTotal        int64  `json:"durationTotal"`
	DurationMin          int64  `json:"durationMin"`
	DurationMax          int64  `json:"durationMax"`
	ServiceDurationTotal int64  `json:"serviceDurationTotal"`
	ServiceDurationMin   int64  `json:"serviceDurationMin"`
	ServiceDurationMax   int64  `json:"serviceDurationMax"`
}
