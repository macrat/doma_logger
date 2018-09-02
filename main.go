package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	temphumid, _ := NewHDC1000Sensor(0x40, 1)
	defer temphumid.Close()

	ss := SensorSet{
		DummySensor([]SensorValue{
			{Name: "dummy_value", Value: 42},
			{Name: "dummy_number", Value: 0.2, Labels: Labels{"type": "a"}},
			{Name: "dummy_number", Value: 0.8, Labels: Labels{"type": "b"}},
		}),
		temphumid,
	}

	fr, _ := NewFluentReporter("localhost", 24224, "doma.alpha.test")
	defer fr.Close()
	rs := ReportServer{
		{Reporter: WriterReporter{os.Stdout}, Interval: 5 * time.Second},
		{Reporter: fr, Interval: 5 * time.Second},
	}

	pe := PrometheusExporter{"doma", ss}

	go pe.ServeForever(":9990", func(err error) {
		fmt.Println(err.Error())
	})
	fmt.Println("running")
	rs.ServeForever(ss, func(err error) {
		fmt.Println(err.Error())
	})
}
