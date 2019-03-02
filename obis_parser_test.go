package main

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

const invalidMsg = `
.0*255(228.3*V)
1-0:96.5.0*255(001C0104)
0-0:96.8.0*255(00F1AD31)
!
`
const singleValidMsg = `
/EZB23092834

1-0:0.0.0*255(1EBZ0100183277)
1-0:96.1.0*255(1EBZ0100183277)
1-0:1.8.0*255(002236.29107286*kWh)
1-0:16.7.0*255(000550.25*W)
1-0:36.7.0*255(000322.14*W)
1-0:56.7.0*255(000052.55*W)
1-0:76.7.0*255(000175.56*W)
1-0:32.7.0*255(227.6*V)
1-0:52.7.0*255(229.3*V)
1-0:72.7.0*255(228.3*V)
1-0:96.5.0*255(001C0104)
0-0:96.8.0*255(00F1AD31)
!
`

func TestObisParser_ParseValid(t *testing.T) {
	as := assert.New(t)

	parser := NewObisParser(context.Background())

	wg := sync.WaitGroup{}
	wg.Add(1)

	parser.Parse(bytes.NewReader([]byte(singleValidMsg)), func(measurement Measurement, err error) {
		as.NoError(err)
		as.Equal(2236.29107286, measurement.TotalKwh)
		as.Equal(550.25, measurement.PTotal)
		as.Equal(322.14, measurement.P1)
		as.Equal(52.55, measurement.P2)
		as.Equal(175.56, measurement.P3)
		as.Equal(227.6, measurement.V1)
		as.Equal(229.3, measurement.V2)
		as.Equal(228.3, measurement.V3)
		wg.Done()
	})

	wg.Wait()
}

func TestObisParser_ParseInvalid(t *testing.T) {
	as := assert.New(t)

	parser := NewObisParser(context.Background())

	wg := sync.WaitGroup{}
	wg.Add(1)

	parser.Parse(bytes.NewReader([]byte(invalidMsg)), func(measurement Measurement, err error) {
		as.Error(err)
		wg.Done()
	})

	wg.Wait()
}
