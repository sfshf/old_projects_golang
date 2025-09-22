package tools

import (
	"time"
)

type Time struct {
	time.Time
}

func Now() Time {
	return Time{time.Now()}
}

func (t Time) String() string {
	if t.Time.IsZero() {
		return "0000-00-00 00:00:00"
	}
	return t.Format("2006-01-02 15:04:05")
}

func (t Time) MarshalJSON() ([]byte, error) {
	if t.Time.IsZero() {
		return []byte(`"0000-00-00 00:00:00"`), nil
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return nil, err
	}
	return []byte(`"` + t.In(loc).Format("2006-01-02 15:04:05") + `"`), nil
}

func (t *Time) UnmarshalJSON(data []byte) error {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return err
	}
	nt, err := time.ParseInLocation(`"2006-01-02 15:04:05"`, string(data), loc)
	if err != nil {
		return err
	}
	*t = Time{Time: nt}
	return nil
}
