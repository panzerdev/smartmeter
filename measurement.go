package main

import "time"

type MeasurementProcessor func(m Measurement, err error)

type Measurement struct {
	Created       time.Time `db:"created_at"`
	MeterID       int       `db:"meter_id"`
	TotalKwhNeg   float64   `db:"total_kwh_neg"`
	TotalKwhPos   float64   `db:"total_kwh_pos"`
	TotalT1KwhPos float64   `db:"total_kwh_t1_pos"`
	TotalT2KwhPos float64   `db:"total_kwh_t2_pos"`
	PTotal        float64   `db:"total_p"`
	P1            float64   `db:"p1"`
	P2            float64   `db:"p2"`
	P3            float64   `db:"p3"`
	V1            float64   `db:"v1"`
	V2            float64   `db:"v2"`
	V3            float64   `db:"v3"`
}
