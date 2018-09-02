package main

import (
	"fmt"
	"io"
	"time"

	"github.com/fluent/fluent-logger-golang/fluent"
)

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
	data := map[string]map[string]interface{}{}

	for _, v := range values {
		data[v.Name] = map[string]interface{}{
			"value":     v.Value,
			"labels":    (map[string]string)(v.Labels),
			"timestamp": v.Timestamp.Format("2006-01-02 15:04:05"),
		}
	}

	return fr.logger.PostWithTime(fr.Tag, time.Now(), data)
}
