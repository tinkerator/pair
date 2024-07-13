// Package pair provides a package abstraction for querying a
// PurpleAir sensor over the local network.
package pair

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// Status is the structure returned by the PurpleAir sensors, via the
// URL path: /json?live=true
type Status struct {
	SensorId           string
	DateTime           string
	Geo                string
	Mem                int
	MemFrag            int     `json:"memfrag"`
	MemFB              int     `json:"memfb"`
	MemCS              int     `json:"memcs"`
	ID                 int     `json:"Id"`
	Lat                float64 `json:"lat"`
	Lon                float64 `json:"lon"`
	Adc                float64
	LoggingRate        int     `json:"loggingrate"`
	Place              string  `json:"place"`
	Version            string  `json:"version"`
	Uptime             int     `json:"uptime"`
	RSSI               int     `json:"rssi"`
	Period             int     `json:"period"`
	HttpSuccess        int     `json:"httpsuccess"`
	HttpSends          int     `json:"httpsends"`
	HardwareVersion    string  `json:"hardwareversion"`
	HardwareDiscovered string  `json:"hardwarediscovered"`
	CurrentTempF       int     `json:"current_temp_f"`
	CurrentHumidity    int     `json:"current_humidity"`
	CurrentDewpointF   int     `json:"current_dewpoint_f"`
	Pressure           float64 `json:"pressure"`
	P24AqicB           string  `json:"p25aqic_b"`
	PM25AqiB           int     `json:"pm2.5_aqi_b"`
	PM10Cf1B           float64 `json:"pm1_0_cf_1_b"`
	PM03UmB            float64 `json:"p_0_3_um_b"`
	PM05UmB            float64 `json:"pm2_5_cf_1_b"`
	P05UmB             float64 `json:"p_0_5_um_b"`
	PM100Cf1B          float64 `json:"pm10_0_cf_1_b"`
	P10UmB             float64 `json:"p_1_0_um_b"`
	PM10AtmB           float64 `json:"pm1_0_atm_b"`
	P25UmB             float64 `json:"p_2_5_um_b"`
	PM25AtmB           float64 `json:"pm2_5_atm_b"`
	P50UmB             float64 `json:"p_5_0_um_b"`
	PM100AtmB          float64 `json:"pm10_0_atm_b"`
	P100UmB            float64 `json:"p_10_0_um_b"`
	P25Aqic            string  `json:"p25aqic"`
	PM25Aqi            int     `json:"pm2.5_aqi"`
	PM10Cf1            float64 `json:"pm1_0_cf_1"`
	P03Um              float64 `json:"p_0_3_um"`
	PM25Cf1            float64 `json:"pm2_5_cf_1"`
	P05Um              float64 `json:"p_0_5_um"`
	PM100Cf1           float64 `json:"pm10_0_cf_1"`
	P10Um              float64 `json:"p_1_0_um"`
	PM10Atm            float64 `json:"pm1_0_atm"`
	P25Um              float64 `json:"p_2_5_um"`
	PM25Atm            float64 `json:"pm2_5_atm"`
	P50Um              float64 `json:"p_5_0_um"`
	PM100Atm           float64 `json:"pm10_0_atm"`
	P100Um             float64 `json:"p_10_0_um"`
	PaLatency          int     `json:"pa_latency"`
	Response           int     `json:"response"`
	ResponseDate       int     `json:"response_date"`
	Latency            int     `json:"latency"`
	WlState            string  `json:"wlstate"`
	Status0            int     `json:"status_0"`
	Status1            int     `json:"status_1"`
	Status2            int     `json:"status_2"`
	Status3            int     `json:"status_3"`
	Status4            int     `json:"status_4"`
	Status6            int     `json:"status_6"`
	SSID               string  `json:"ssid"`
}

// Sensor holds a cached summary of a PurpleAir sensor state.
type Sensor struct {
	Addr       string
	mu         sync.Mutex
	LastSample *Status

	// Polynomial best fit for temperature conversion.
	TempPoly []float64
}

// NewSensor registers a new sensor reference. This cannot fail, it
// does not attempt to sample the device. Use the Refresh method.
func NewSensor(addr string) *Sensor {
	return &Sensor{
		Addr: addr,
	}
}

// TempAdjust sets the polynomial expansion parameters to convert raw
// temperature measurements to calibrated values. This adjustment is
// applied every time one of the (*Sensor) temperature() functions are
// called. The default is to not adjust the values, i.e. 1:1.
//
// For reference, the sensor this code was developed for, fitting a
// polynomial to readings of a digital thermometer placed next to it,
// we use the following coefficients: {-8.9037,1.0441}. This is a
// linear fit with the raw values are 9F too high, and the individual
// raw F measurements appear to be about 4% smaller than a calibrated
// F unit.
func (s *Sensor) TempAdjust(coef []float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TempPoly = coef
}

// expand a temperature measurement with the TempPoly coefficients.
func actualTemp(coef []float64, raw float64) float64 {
	if len(coef) == 0 {
		return raw
	}
	t, x := 0.0, 1.0
	for _, c := range coef {
		t += c * x
		x *= raw
	}
	return t
}

// CtoF is the Celsius to Fahrenheit conversion.
func CtoF(c float64) float64 {
	return 9*c/5 + 32
}

// FtoC is the Fahrenheit to Celsius conversion.
func FtoC(c float64) float64 {
	return (c - 32) * 5 / 9
}

// Temp returns the adjusted temperature value. Depending on how the
// AdjustTemp value is calibrated, the units of the returned value
// very dependent on that calibration choice. Convenience functions
// for CtoF and FtoC are provided.
func (s *Sensor) Temp() float64 {
	s.mu.Lock()
	t := s.LastSample.CurrentTempF
	cs := s.TempPoly
	s.mu.Unlock()
	return actualTemp(cs, float64(t))
}

// DewPoint returns the current dew point temerature.
func (s *Sensor) DewPoint() float64 {
	s.mu.Lock()
	t := s.LastSample.CurrentDewpointF
	cs := s.TempPoly
	s.mu.Unlock()
	return actualTemp(cs, float64(t))
}

// Humidity returns the percent humidity seen by the sensor.
func (s *Sensor) Humidity() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return float64(s.LastSample.CurrentHumidity)
}

// Pressure returns the current Pressure in hPa units.
func (s *Sensor) Pressure() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.LastSample.Pressure
}

// AQIA returns the AQI (Air Quality Index) value for sensor A.
func (s *Sensor) AQIA() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return float64(s.LastSample.PM25Aqi)
}

// AQIB returns the AQI (Air Quality Index) value for sensor B.
func (s *Sensor) AQIB() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return float64(s.LastSample.PM25AqiB)
}

// Refresh fetches and updates the cached Sensor state.
func (s *Sensor) Refresh() error {
	resp, err := http.Get(fmt.Sprint("http://", s.Addr, "/json?live=true"))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	status := &Status{}
	if err := dec.Decode(status); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastSample = status
	return nil
}
