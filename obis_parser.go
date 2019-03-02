package main

import (
	"bufio"
	"context"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	OBISRegex  = `(?P<obis>\d-\d:\d{1,2}\.\d{1,2}\.\d\*\d{3})(?P<value>\(\S*\))`
	ValueRegex = `(\d*\.\d*)\*(\S[^\)]*)`
)

const (
	OBISMsgSeparator = "!"
	OBIScodeCurrent  = "1-0:1.8.0*255"  // 2248.01818051 kWh
	OBIScodePt       = "1-0:16.7.0*255" // 282.03 W
	OBIScodeP1       = "1-0:36.7.0*255" // 63.33 W
	OBIScodeP2       = "1-0:56.7.0*255" // 30.14 W
	OBIScodeP3       = "1-0:76.7.0*255" // 188.56 W
	OBIScodeV1       = "1-0:32.7.0*255" // 233.8 V
	OBIScodeV2       = "1-0:52.7.0*255" // 236.1 V
	OBIScodeV3       = "1-0:72.7.0*255" // 235.1 V
)

var (
	obisExpress  = regexp.MustCompile(OBISRegex)
	valueExpress = regexp.MustCompile(ValueRegex)
)

type ObisParser struct {
	msgSeparator  string
	lineValidator *regexp.Regexp
	valueParser   *regexp.Regexp
	ctx           context.Context
}

func NewObisParser(ctx context.Context) *ObisParser {
	return &ObisParser{
		msgSeparator:  OBISMsgSeparator,
		lineValidator: obisExpress,
		valueParser:   valueExpress,
		ctx:           ctx,
	}
}

// reader is read until io.EOF
//
// MeasurementProcessor is called in a new Goroutine
func (parser ObisParser) Parse(reader io.Reader, mp MeasurementProcessor) {
	scanner := bufio.NewScanner(reader)
	var msg []string
	for scanner.Scan() {
		line := scanner.Text()
		if line == parser.msgSeparator {
			measurement, err := parseObis(msg)
			go mp(measurement, err)

			msg = []string{}
		} else if obisExpress.MatchString(line) {
			msg = append(msg, line)
		}

		select {
		case <-parser.ctx.Done():
			return
		default:
		}
	}
}

func parseObis(msg []string) (Measurement, error) {
	if len(msg) != 12 {
		return Measurement{}, errors.Errorf("\nReading has only size %v of 12:\n%v", len(msg), strings.Join(msg, "\n"))
	}

	measurement := Measurement{
		Created: time.Now(),
		MeterID: 1, // use a real id if you have more then 1 meter
	}

	for _, v := range msg {
		matches := obisExpress.FindStringSubmatch(v)[1:]
		if len(matches) < 2 {
			continue
		}

		key := matches[0]
		value := matches[1]
		if valueExpress.MatchString(value) {
			values := valueExpress.FindStringSubmatch(value)[1:]
			num, err := strconv.ParseFloat(values[0], 64)
			if err != nil {
				log.Println("Parse float", err)
				continue
			}
			mapValues(key, num, &measurement)
		}
	}

	return measurement, nil
}

func mapValues(obis string, value float64, m *Measurement) {
	switch obis {
	case OBIScodeCurrent:
		m.TotalKwh = value
	case OBIScodePt:
		m.PTotal = value
	case OBIScodeP1:
		m.P1 = value
	case OBIScodeP2:
		m.P2 = value
	case OBIScodeP3:
		m.P3 = value
	case OBIScodeV1:
		m.V1 = value
	case OBIScodeV2:
		m.V2 = value
	case OBIScodeV3:
		m.V3 = value
	}
}
