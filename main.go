package main

import (
	"fmt"
	"time"
)

func main() {
	temphumid, _ := NewHDC1000Sensor(0x40, 1)
	defer temphumid.Close()

	fr, _ := NewFluentReporter("localhost", 24224, "doma.alpha.test")
	defer fr.Close()

	go PrometheusExporter{"doma", temphumid}.ServeForever(":9990", func(err error) {
		fmt.Println(err.Error())
	})

	ReportServer{
		{Reporter: fr, Interval: 5 * time.Second},
	}.ServeForever(temphumid, func(err error) {
		fmt.Println(err.Error())
	})
}
