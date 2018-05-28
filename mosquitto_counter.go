package main

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
)

// MosquittoCounter exports all counter metrics are already added by mosquitto
type MosquittoCounter struct {
	Desc *prometheus.Desc
	counter
	// ... many more fields
}

// NewMosquittoCounter get a new one
func NewMosquittoCounter(desc *prometheus.Desc) *MosquittoCounter {
	return &MosquittoCounter{
		Desc: desc,
	}
}

// Set sets the value
func (c *MosquittoCounter) Set(v float64) {
	c.counter.Set(v)
}

// Describe simply sends the two Descs in the struct to the channel.
func (c *MosquittoCounter) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Desc
}

// Collect already added counter values
func (c *MosquittoCounter) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		c.Desc,
		prometheus.CounterValue,
		c.counter.value,
	)
}

type counter struct {
	value float64
}

func (c *counter) Set(v float64) {
	if v < 0 {
		panic(errors.New("counter cannot decrease in value"))
	}
	c.value = v
}
