package main

import (
	"time"
	"strings"
	"encoding/binary"

	"github.com/d2r2/go-i2c"
)

type Sensor interface {
	Read() ([]SensorValue, error)
}

type DummySensor []SensorValue

func (ds DummySensor) Read() ([]SensorValue, error) {
	for k := range ds {
		ds[k].Timestamp = time.Now()
	}
	return ([]SensorValue)(ds), nil
}

type HDC1000Sensor struct {
	Prefix string
	bus    *i2c.I2C
}

func NewHDC1000Sensor(address uint8, bus int) (h HDC1000Sensor, err error) {
	h.bus, err = i2c.NewI2C(address, bus)

	return h, err
}

func (h HDC1000Sensor) Close() {
	h.bus.Close()
}

func (h HDC1000Sensor) Read() ([]SensorValue, error) {
	if _, err := h.bus.WriteBytes([]byte{0x00}); err != nil {
		return nil, err
	}

	time.Sleep(13 * time.Millisecond)

	data := make([]byte, 4)
	if _, err := h.bus.ReadBytes(data); err != nil {
		return nil, err
	}

	timestamp := time.Now()
	prefix := h.Prefix
	if prefix != "" && !strings.HasSuffix(prefix, "_") {
		prefix += "_"
	}

	return []SensorValue{
		{
			Name: prefix + "temperature",
			Value: float64(binary.BigEndian.Uint16(data[:2])) / float64(0xFFFF) * 165.0 - 40.0,
			Timestamp: timestamp,
		},
		{
			Name: prefix + "humidity",
			Value: float64(binary.BigEndian.Uint16(data[2:])) / float64(0xFFFF),
			Timestamp: timestamp,
		},
	}, nil
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
