package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fluent/fluent-logger-golang/fluent"
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

type Sensor interface {
	Read() ([]SensorValue, error)
}

type DummySensor []SensorValue

func (ds DummySensor) Read() ([]SensorValue, error) {
	for k, _ := range ds {
		ds[k].Timestamp = time.Now()
	}
	return ([]SensorValue)(ds), nil
}

type SensorSet []Sensor

func (ss SensorSet) Read() ([]SensorValue, error) {
	var result []SensorValue

	for _, s := range ss {
		r, err := s.Read()
		if err != nil {
			return nil, err
		}

		for _, x := range r {
			result = append(result, x)
		}
	}

	return result, nil
}

type Reporter interface {
	Report(values []SensorValue) error
}

type WriterReporter struct {
	Writer io.Writer
}

func (wr WriterReporter) Report(values []SensorValue) error {
	for _, x := range values {
		fmt.Fprintf(wr.Writer, " %s%s=%f", x.Name, x.Labels, x.Value)
	}
	fmt.Fprintln(wr.Writer, "")

	return nil
}

type FluentReporter struct {
	logger *fluent.Fluent
	Tag    string
}

func NewFluentReporter(host string, port int, tag string) (fr FluentReporter, err error) {
	fr.Tag = tag

	fr.logger, err = fluent.New(fluent.Config{
		FluentHost: host,
		FluentPort: port,
	})
	return fr, err
}

func (fr FluentReporter) Close() {
	fr.logger.Close()
}

func (fr FluentReporter) Report(values []SensorValue) error {
	data := map[string]map[string]interface{} {}

	for _, v := range values {
		data[v.Name] = map[string] interface{} {
			"value": v.Value,
			"labels": (map[string]string)(v.Labels),
			"timestamp": v.Timestamp.Format("2006-01-02 15:04:05"),
		}
	}

	return fr.logger.PostWithTime(fr.Tag, time.Now(), data)
}

type ReportRequest struct {
	Reporter     Reporter
	Interval     time.Duration
	lastReported time.Time
}

type Reporters []*ReportRequest

func (rs Reporters) Report(values []SensorValue) error {
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

func (rs Reporters) NextReportingTime() (next time.Time) {
	next = time.Unix(1<<63-1, 0)

	for _, r := range rs {
		if t := r.lastReported.Add(r.Interval); next.Before(t) {
			next = t
		}
	}

	return
}

func (rs Reporters) Serve(sensor Sensor) error {
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

func (rs Reporters) ServeForever(sensor Sensor, onError func(error)) {
	for {
		time.Sleep(rs.NextReportingTime().Sub(time.Now()))

		if values, err := sensor.Read(); err != nil {
			onError(err)
		} else if err = rs.Report(values); err != nil {
			onError(err)
		}
	}
}

type PrometheusExporter struct {
	Prefix  string
	Sensors SensorSet
}

func (pe PrometheusExporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	values, err := pe.Sensors.Read()
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, err.Error())
		return
	}

	for _, v := range values {
		fmt.Fprintf(w, "%s_%s%s %f %d\n", pe.Prefix, v.Name, v.Labels, v.Value, v.Timestamp.Unix())
	}
}

func (pe PrometheusExporter) Serve(addr string) error {
	http.Handle("/metrics", pe)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "<a href=\"/metrics\">metrics</a>")
	})
	return http.ListenAndServe(addr, nil)
}

func (pe PrometheusExporter) ServeForever(addr string, onError func(error)) {
	for {
		onError(pe.Serve(addr))
	}
}

func main() {
	ss := SensorSet{DummySensor([]SensorValue{
		{Name: "dummy_value", Value: 42},
		{Name: "dummy_number", Value: 0.2, Labels: Labels{"type": "a"}},
		{Name: "dummy_number", Value: 0.8, Labels: Labels{"type": "b"}},
	})}

	fr, _ := NewFluentReporter("localhost", 24224, "doma.sensors")
	defer fr.Close()
	rs := Reporters{
		{Reporter: WriterReporter{os.Stdout}, Interval: 5 * time.Second},
		{Reporter: fr, Interval: 5 * time.Second},
	}

	pe := PrometheusExporter{"doma", ss}

	go pe.ServeForever(":8888", func(err error) {
		fmt.Println(err.Error())
	})
	rs.ServeForever(ss, func(err error) {
		fmt.Println(err.Error())
	})
}
