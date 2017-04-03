package main

/*
	setup
	docker run -i -p 3000:3000 grafana/grafana
	import dashboard
	modify prometheus.yaml to point to current ip (docker localhost is container-only)
	docker run -p 9090:9090 -v $PWD/prometheus.yml:/etc/prometheus/prometheus.yml prom/prometheus
	run scale
 */

import (
	"github.com/GeertJohan/go.hid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"math"
	"net/http"
	"time"
)

const KG_MODE = 3

type Scale struct {
	device *hid.Device
}

var (
	catWent = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cat_went",
		Help: "Cat used the litter box",
	})
	currentCatWeight = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cat_weight",
		Help: "Last recorded cat weight in pounds",
	})
	catExcrement = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cat_excrement",
		Help: "Amount of exrement left over time in pounds",
	})
	scoops = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "scoops",
		Help: "The litter box was scooped",
	})
	emptied = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "emptied",
		Help: "The litter box was emptied",
	})
	refilled = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "refilled",
		Help: "The litter box was refilled",
	})
	litterAdd = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "litter_added",
		Help: "More litter was added",
	})
	currentWeight = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "current_weight",
		Help: "The current weight of the litter box with contents in pounds",
	})
)

func main() {
	println("Starting....")
	prometheus.MustRegister(catWent, currentCatWeight, catExcrement, scoops, emptied, refilled, litterAdd, currentWeight)
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":8005", nil)
	weightEvents := StartWeightListener()
	stableEvents := StartStableWeightListener(weightEvents, 3*time.Second)
	litterBoxEvents := StartLitterBoxListener(stableEvents)
	for {
		event := <-litterBoxEvents
		log.Println(event)
	}

	/* todo:
	graph changes over time, using WeightListener, a database, and some sort of graphing thing
		grafana, with influxdb, elasticsearch, or graphite
	https://godoc.org/github.com/prometheus/client_golang/prometheus

	*/

}

func InitScale() Scale {
	// requires hidapi from brew
	device, err := hid.Open(0x922, 0x8009, "")
	if err != nil {
		log.Fatal("error connecting to scale: ", err)
	}
	return Scale{device}
}

func (s Scale) Close() {
	s.device.Close()
}

// Returns current scale reading, in pounds. Converts units to pounds if scale is in kilogram mode.
func (s Scale) Read() float64 {
	bytes := make([]byte, 6)
	len, err := s.device.Read(bytes)
	if err != nil || len != 6 {
		log.Fatal("Coudln't read from scale: ", err)
	}
	weight := (float64(bytes[4]) + float64(bytes[5])*256) / 10
	if bytes[2] == KG_MODE {
		weight = math.Floor(100*weight*2.20462+.5) / 100
	}
	return weight
}
