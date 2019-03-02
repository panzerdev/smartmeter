package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	flag "github.com/spf13/pflag"
	"go.bug.st/serial.v1"
)

var (
	flagPersistence = flag.String("persist", "none", "type of persistence: [none, postgres]")

	dbHost        = flag.String("db-host", "localhost", "Db host")
	dbPort        = flag.String("db-port", "5432", "Db port")
	dbUser        = flag.String("db-user", "postgres", "Db user")
	dbPassword    = flag.String("db-password", "root", "Db password")
	dbName        = flag.String("db-name", "root", "Db name")
	flushInterval = flag.Duration("db-flush", time.Minute, "Flush after duration")

	usbDevName = flag.String("reader-port", "/dev/ttyUSB0", "Device name of reader")
)

var (
	collector = make(chan Measurement)
	retry     = make(chan []Measurement)
)

func main() {
	log.Println("Starting... ")
	flag.Parse()
	flag.VisitAll(func(f *flag.Flag) {
		log.Println("Flag:", f.Name, f.DefValue, f.Value, ":::", f.Usage)
	})

	initProm()

	ctx, c := context.WithCancel(context.Background())
	defer c()

	// TODO make this configurable
	mode := &serial.Mode{
		BaudRate: 9600,
		DataBits: 7,
		Parity:   serial.EvenParity,
		StopBits: serial.OneStopBit,
	}
	file, err := serial.Open(*usbDevName, mode)
	if err != nil {
		log.Fatal("Error open", err)
		time.Sleep(time.Second * 5)
	}
	defer file.Close()

	parser := NewObisParser(ctx)
	go parser.Parse(file, handleMeasurement)

	go func() {
		// Start server for serving prometheus
		log.Fatalln(http.ListenAndServe(":8080", nil))
	}()

	go func() {
		signalC := make(chan os.Signal, 1)
		signal.Notify(signalC, syscall.SIGTERM)
		<-signalC
		log.Println("Shutdown received")
		c()
	}()

	startCollector(ctx, getWriter(*flagPersistence))
}

func getWriter(option string) MeasurementPersister {
	switch option {
	case "none":
		return NoOpWriter(true)
	case "postgres":
		return NewPostgresWriter(PostgresConfig{
			User:     *dbUser,
			Password: *dbPassword,
			Host:     *dbHost,
			Port:     *dbPort,
			DbName:   *dbName,
		})
	default:
		log.Fatalf("Unkown persist flag: %v of [none, postgres]\n", option)
	}
	return nil
}

func startCollector(ctx context.Context, persister MeasurementPersister) {
	var measurements []Measurement
	var wg sync.WaitGroup

	ticker := time.NewTicker(*flushInterval)
	defer ticker.Stop()

	for {
		select {
		case m := <-collector:
			measurements = append(measurements, m)
		case m := <-retry:
			measurements = append(measurements, m...)
		case <-ticker.C:
			wg.Add(1)
			go func(m []Measurement) {
				defer wg.Done()
				err := persister.Flush(m)
				if err != nil {
					log.Println(err)
					return
				}
			}(measurements)

			measurements = []Measurement{}
		case <-ctx.Done():
			// TODO needed if flush call in sync?
			err := persister.Flush(measurements)
			if err != nil {
				log.Println(err)
			}

			wg.Wait()
			log.Println("Done with all things.. bye bye")
			return
		}
	}
}

func handleMeasurement(m Measurement, err error) {
	if err != nil {
		log.Println(err)
		return
	}

	meter(m)
	collector <- m
}
