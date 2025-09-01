package entity

import "time"

type MinuteStat struct {
	Ts time.Time
	V  int64
}
