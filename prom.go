package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	voltage    *prometheus.GaugeVec
	power      *prometheus.GaugeVec
	totalPower prometheus.Gauge
	total      prometheus.Gauge
)

const (
	Phase = "phase"
	Unit  = "unit"
	Obis  = "obis"
)

func initProm() {
	voltage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "smart_meter_voltage",
		Help: "Current Voltage of each phase",
		ConstLabels: prometheus.Labels{
			Unit: "V",
		},
	}, []string{
		Phase,
		Obis,
	})
	power = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "smart_meter_power",
		Help: "Current power of each phase",
		ConstLabels: prometheus.Labels{
			Unit: "W",
		},
	}, []string{
		Phase,
		Obis,
	})

	total = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "smart_meter_current_total_reading",
		Help: "Current total value ",
		ConstLabels: prometheus.Labels{
			Unit: "kWh",
			Obis: OBIScodeCurrent},
	})

	totalPower = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "smart_meter_current_total_power",
		Help: "Current total power for all phases",
		ConstLabels: prometheus.Labels{
			Unit: "W",
			Obis: OBIScodePt},
	})

	prometheus.MustRegister(voltage)
	prometheus.MustRegister(power)
	prometheus.MustRegister(total)
	prometheus.MustRegister(totalPower)

	http.Handle("/metrics", promhttp.Handler())
}

func meter(m Measurement) {
	total.Set(m.TotalKwh)
	totalPower.Set(m.PTotal)
	power.WithLabelValues("all", OBIScodePt).Set(m.PTotal)
	power.WithLabelValues("1", OBIScodeP1).Set(m.P1)
	power.WithLabelValues("2", OBIScodeP2).Set(m.P2)
	power.WithLabelValues("3", OBIScodeP3).Set(m.P3)
	voltage.WithLabelValues("1", OBIScodeV1).Set(m.V1)
	voltage.WithLabelValues("2", OBIScodeV2).Set(m.V2)
	voltage.WithLabelValues("3", OBIScodeV3).Set(m.V3)
}
