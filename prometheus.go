package main

import (
	"fmt"
	"net/http"
)

type PrometheusExporter struct {
	Prefix string
	Sensor Sensor
}

func (pe PrometheusExporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	values, err := pe.Sensor.Read()
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
