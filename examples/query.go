package main

import (
	"flag"
	"log"
	"strconv"
	"strings"
	"time"

	"zappem.net/pub/io/pair"
)

// Program query performs a local query to a PurpleAir sensor and
// displays the current sensor measurements.

var (
	addr  = flag.String("sensor", "", "local network address of sensor")
	poll  = flag.Duration("poll", 0*time.Second, "non-zero polls with this interval")
	retry = flag.Int("retry", 3, "default number of times to retry request - once a second")
	coef  = flag.String("coef", "-8.9037,1.0441", "comma separated coefficients for temperature conversion")
)

func main() {
	flag.Parse()

	if *addr == "" {
		log.Fatal("--sensor <net-address>, is required")
	}
	s := pair.NewSensor(*addr)
	if *coef != "" {
		var cs []float64
		for _, c := range strings.Split(*coef, ",") {
			x, err := strconv.ParseFloat(c, 64)
			if err != nil {
				log.Fatalf("failed to parse %q (from %q): %v", c, *coef, err)
			}
			cs = append(cs, x)
		}
		s.TempAdjust(cs)
	}
	retries := *retry
	for {
		if err := s.Refresh(); err != nil {
			if retries > 0 {
				retries--
				time.Sleep(1 * time.Second)
				continue
			}
			log.Fatalf("failed to refresh: %v (after %d tries)", err, *retry)
		}
		t := s.Temp()
		tC := pair.FtoC(t)
		dew := s.DewPoint()
		dewC := pair.FtoC(dew)
		pres := s.Pressure()
		hum := s.Humidity()
		aqiA := s.AQIA()
		aqiB := s.AQIB()
		log.Printf("temp=%.1fF(%.1fC) dewPt=%.1fF(%.1fC) hum=%g%% pres=%.1fhPa AQIab=%.1f,%.1f", t, tC, dew, dewC, hum, pres, aqiA, aqiB)
		if *poll == 0 {
			break
		}
		time.Sleep(*poll)
	}
}
