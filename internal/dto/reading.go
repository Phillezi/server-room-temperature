package dto

import "time"

type Reading struct {
	Timestamp  time.Time `json:"timestamp"`
	TempMilliC int32     `json:"temp_millic"`
}
