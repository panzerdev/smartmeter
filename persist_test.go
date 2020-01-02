package main

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestPostgresWriter_Flush(t *testing.T) {
	rand.Seed(time.Now().Unix())
	meterId := rand.Intn(10000)
	var m []Measurement
	createdAt := time.Now()
	for i := 0; i < 100; i++ {
		m = append(m, Measurement{
			Created:       createdAt.Add(time.Microsecond * time.Duration(i)),
			MeterID:       meterId,
			TotalKwhNeg:   rand.Float64(),
			TotalKwhPos:   rand.Float64(),
			TotalT1KwhPos: rand.Float64(),
			TotalT2KwhPos: rand.Float64(),
			PTotal:        rand.Float64(),
			P1:            rand.Float64(),
			P2:            rand.Float64(),
			P3:            rand.Float64(),
			V1:            rand.Float64(),
			V2:            rand.Float64(),
			V3:            rand.Float64(),
		})
	}

	writer := NewPostgresWriter(PostgresConfig{
		User:     "postgres",
		Password: "root",
		Host:     "localhost",
		Port:     "5432",
		DbName:   "db",
	})
	_, err := writer.db.Exec("DELETE from meter_data where meter_id = $1", meterId)
	assert.NoError(t, err)

	err = writer.Flush(m)

	assert.NoError(t, err)

	var nr int
	err = writer.db.Get(&nr, "SELECT count(*) from meter_data where meter_id = $1;", meterId)
	assert.NoError(t, err)
	assert.Equal(t, len(m), nr)

	_, err = writer.db.Exec("DELETE from meter_data where meter_id = $1", meterId)
	assert.NoError(t, err)
}
