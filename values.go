package main

import (
	"fmt"
	"strings"
	"time"
)

type Labels map[string]string

func (ls Labels) String() string {
	if len(ls) == 0 {
		return ""
	}

	var rs []string

	for k, v := range ls {
		rs = append(rs, fmt.Sprintf("%s=\"%s\"", k, v))
	}

	return "{" + strings.Join(rs, ",") + "}"
}

type SensorValue struct {
	Name      string
	Labels    Labels
	Value     float64
	Timestamp time.Time
}
