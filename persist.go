package main

import (
	"fmt"
	"log"
	"sort"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type MeasurementPersister interface {
	Flush(measurements []Measurement) error
}

type NoOpWriter bool

func (w NoOpWriter) Flush(measurements []Measurement) error {
	if w {
		log.Printf("Not persisting %v items\n", len(measurements))
	}
	return nil
}

type PostgresWriter struct {
	retryC chan []Measurement
	db     *sqlx.DB
}

//	CREATE TABLE meter_data (
//		created_at 	timestamp PRIMARY KEY NOT NULL,
//		meter_id 	integer NOT NULL,
//		total_kwh 	decimal NOT NULL,
//		total_p 	decimal NOT NULL,
//		p1 			decimal NOT NULL,
//		p2 			decimal NOT NULL,
//		p3 			decimal NOT NULL,
//		v1 			decimal NOT NULL,
//		v2 			decimal NOT NULL,
//		v3 			decimal NOT NULL
//	);

const insert_meter_data = `INSERT INTO meter_data( created_at, meter_id, total_kwh, total_p, p1, p2, p3, v1, v2, v3) 
								VALUES(:created_at, :meter_id, :total_kwh, :total_p, :p1, :p2, :p3, :v1, :v2, :v3);`

type PostgresConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	DbName   string
}

func (pc PostgresConfig) String() string {
	return fmt.Sprintf(
		`user=%v password=%v host=%v port=%v dbname=%v sslmode=disable`,
		pc.User, pc.Password, pc.Host, pc.Port, pc.DbName)
}

func NewPostgresWriter(conf PostgresConfig) *PostgresWriter {
	db := sqlx.MustConnect("postgres", conf.String())
	return &PostgresWriter{
		retryC: retry,
		db:     db,
	}
}

func (p *PostgresWriter) Flush(measurements []Measurement) error {
	// TODO flush to DB and write to retry chan on error

	sort.Slice(measurements, func(i, j int) bool {
		return measurements[i].Created.Before(measurements[j].Created)
	})

	exec := func() error {
		tx, err := p.db.Beginx()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		insert, err := tx.PrepareNamed(insert_meter_data)
		if err != nil {
			return err
		}

		for _, m := range measurements {
			_, err := insert.Exec(m)
			if err != nil {
				return err
			}
		}

		return tx.Commit()
	}()

	if err := exec; err != nil {
		log.Println(err)
		go func() { p.retryC <- measurements }()
		return err
	}
	return nil
}
