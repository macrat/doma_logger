package main

import (
	"time"
)

type ReportConfig struct {
	Reporter     Reporter
	Interval     time.Duration
	lastReported time.Time
}

type ReportServer []*ReportConfig

func (rs ReportServer) Report(values []SensorValue) error {
	for _, r := range rs {
		if time.Now().Sub(r.lastReported) >= r.Interval {
			r.lastReported = time.Now()
			if err := r.Reporter.Report(values); err != nil {
				return err
			}
		}
	}

	return nil
}

func (rs ReportServer) NextReportingTime() (next time.Time) {
	next = time.Unix(1<<63-1, 0)

	for _, r := range rs {
		if t := r.lastReported.Add(r.Interval); next.Before(t) {
			next = t
		}
	}

	return
}

func (rs ReportServer) Serve(sensor Sensor) error {
	for {
		time.Sleep(rs.NextReportingTime().Sub(time.Now()))

		values, err := sensor.Read()
		if err != nil {
			return err
		}

		if err = rs.Report(values); err != nil {
			return err
		}
	}

	return nil
}

func (rs ReportServer) ServeForever(sensor Sensor, onError func(error)) {
	for {
		time.Sleep(rs.NextReportingTime().Sub(time.Now()))

		if values, err := sensor.Read(); err != nil {
			onError(err)
		} else if err = rs.Report(values); err != nil {
			onError(err)
		}
	}
}
