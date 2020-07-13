package common

import "time"

type Config struct {
	CreateFilePrefix string
	StartTime        time.Time
	Workers          int
}
