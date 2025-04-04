package handlers

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/coldstar-507/media-server/internal/paths"
)

/*
1d, 2d => 15 minutes
1w, 2w, 1M => 1 hour
3M, 1Y, ALL => 1 day
*/

var (
	wocUrl  = "https://api.whatsonchain.com/v1/bsv/main/exchangerate/historical?from=%d&to=%d"
	timeFmt = "2022-11-12T00:43:55.000Z"
	ogUrl   = "https://api.orangegateway.com/graphql"
	minFmt  = "2006-01-02 15:04:05"
)

type TimeRate struct {
	Rate float32 `json:"rate"`
	Time int64   `json:"time"`
}

var PriceAgent = &priceAgent{
	rch: make(chan *priceReadReq),
}

type priceReadReq struct {
	resWriter   io.Writer
	until       int64
	done        chan struct{}
	periodicity string
}
type priceAgent struct {
	rch chan *priceReadReq
}

func makeUrl(from, to int64) string {
	return fmt.Sprintf(wocUrl, from, to)
}

func validPeriodicity(periodicity string) bool {
	switch periodicity {
	case "day", "hour", "minute15", "":
		return true
	default:
		return false
	}
}

func pathOfPeriodicity(periodicity string) string {
	switch periodicity {
	case "day", "":
		return paths.HISTORICAL_RATES_FILE
	case "hour":
		return paths.ONE_HOUR_RATES_FILE
	case "minute15":
		return paths.FIFTEEN_MIN_RATES_FILE
	default:
		panic("pathOfPeriodicity(" + periodicity + ")")
	}
}

/*
1d, 2d => 15 minutes
1w, 2w, 1M => 1 hour
3M, 1Y, ALL => 1 day
*/

func defaultTimeOfPeriodicity(periodicity string) time.Time {
	switch periodicity {
	case "day", "":
		return time.Now().Add(time.Hour * 24 * 365 * 10 * -1)
	case "hour":
		return time.Now().Add(time.Hour * 24 * 30 * 4 * -1)
	case "minute15":
		return time.Now().Add(time.Hour * 24 * 2 * -1)
	default:
		panic("pathOfPeriodicity(" + periodicity + ")")
	}
}

func HandleRatesRequest(w http.ResponseWriter, r *http.Request) {
	untilStr := r.PathValue("until")
	periodicity := r.PathValue("periodicity")
	if !validPeriodicity(periodicity) {
		log.Println("HandleRatesRequest: invalid periodicity:", periodicity)
		w.WriteHeader(500)
	} else if until, err := strconv.ParseInt(untilStr, 10, 64); err != nil {
		log.Println("HandlePriceRequest: error reading until:", err)
		w.WriteHeader(501)
	} else {
		done := make(chan struct{})
		defer close(done)
		PriceAgent.rch <- &priceReadReq{
			resWriter:   w,
			until:       until,
			done:        done,
			periodicity: periodicity,
		}
		<-done
		log.Println("HandleRatesRequest: done")
	}
}

func (pa *priceAgent) Run() {
	ticDay := time.NewTicker(time.Hour * 24)
	ticMin15 := time.NewTicker(time.Minute * 15)
	ticHour := time.NewTicker(time.Hour)
	getLatestPrices("minute15")
	getLatestPrices("hour")
	getLatestPrices("day")
	for {
		log.Println("PriceAgent: waiting for request")
		select {
		case <-ticDay.C:
			getLatestPrices("day")
		case <-ticHour.C:
			getLatestPrices("hour")
		case <-ticMin15.C:
			getLatestPrices("minute15")
		case req := <-pa.rch:
			log.Printf("PriceAgent: Run(): req until=%d\n", req.until)
			err := readUntil(req.resWriter, req.until, req.periodicity)
			if err != nil {
				log.Printf("PriceAgent: Run(): readUnitil(...) error: %v\n", err)
			}
			req.done <- struct{}{}
		}
	}
}

func getLatestPrices(periodicity string) {
	log.Printf("getLatestPrices(%s)\n", periodicity)
	lt, err := latestTime(periodicity)
	if err != nil {
		log.Println("PriceAgent: Run(): latestTime() error:", err)
	}
	err = appendFrom(lt.Unix(), periodicity)
	if err != nil {
		log.Printf("PriceAgent: Run(): appendFrom(%d, %s) error: %v\n",
			lt.Unix(), periodicity, err)
	}
}

func getWocRates(latest int64) ([]TimeRate, error) {
	url := makeUrl(latest, time.Now().Unix())
	fmt.Println(url)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	var j []TimeRate

	err = json.NewDecoder(res.Body).Decode(&j)
	if err != nil {
		return nil, err
	}

	return j, nil
}

func appendFrom(latest int64, periodicity string) error {
	path := pathOfPeriodicity(periodicity)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}

	var rates []TimeRate
	if len(periodicity) == 0 || periodicity == "day" {
		rates, err = getWocRates(latest)
	} else {
		rates, err = getOgFrom(time.Unix(latest, 0), periodicity)
	}
	if err != nil {
		return err
	}

	for _, m := range rates {
		if latest == m.Time {
			fmt.Printf("skipping %d (equals latest)\n", m.Time)
			continue
		}
		fmt.Printf("rate=%f, time=%d\n", m.Rate, m.Time)
		err = binary.Write(f, binary.BigEndian, m.Rate)
		if err != nil {
			return err
		}
		err = binary.Write(f, binary.BigEndian, m.Time)
		if err != nil {
			return err
		}
	}

	return f.Close()
}

func readUntil(w io.Writer, until int64, periodicity string) error {
	// Open the file
	file, err := os.Open(pathOfPeriodicity(periodicity))
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get the file size
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Get the size of the file
	fileSize := fileInfo.Size()

	// Define the block size (12 bytes)
	blockSize := int64(12)

	// Loop to read 12-byte blocks from the end
	var rate float32
	var time int64
	for position := fileSize - blockSize; position >= 0; position -= blockSize {
		// Seek to the position
		_, err := file.Seek(position, 0)
		if err != nil {
			return fmt.Errorf("failed to seek: %w", err)
		}

		// Read the 12-byte block
		if err = binary.Read(file, binary.BigEndian, &rate); err != nil {
			return fmt.Errorf("failed to read rate: %w", err)
		}
		if err = binary.Read(file, binary.BigEndian, &time); err != nil {
			return fmt.Errorf("failed to read time: %w", err)
		}

		if time > until {
			binary.Write(w, binary.BigEndian, rate)
			binary.Write(w, binary.BigEndian, time)
		}
		// Process the block (for now, print it)
		fmt.Printf("Rate/Time at position %d: rate: %f, time: %d\n", position, rate, time)
	}

	return nil

}

func latestTime(periodicity string) (time.Time, error) {
	path := pathOfPeriodicity(periodicity)
	defaultTime := defaultTimeOfPeriodicity(periodicity)
	file, err := os.Open(path)
	if err != nil {
		return defaultTime, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return defaultTime, fmt.Errorf("failed to stat file: %w", err)
	}

	fileSize := fileInfo.Size()

	blockSize := int64(8)
	position := fileSize - blockSize

	_, err = file.Seek(position, 0)
	if err != nil {
		return defaultTime, fmt.Errorf("failed to seek: %w", err)
	}

	var t int64
	if err = binary.Read(file, binary.BigEndian, &t); err != nil {
		return defaultTime, fmt.Errorf("failed to read time: %w", err)
	}

	return time.Unix(t, 0), nil
}

func makeq(from, to time.Time, periodicity string) ([]byte, error) {
	m := map[string]any{
		"query": `query ($instrument_id: String!, $limit: Int, $date_range: DateRangeInput, $periodicity: InstrumentHistoryPeriodicity) {
   instrument_price_bars (instrument_id: $instrument_id, limit: $limit, date_range: $date_range, periodicity: $periodicity) { instrument_id, ts, close }}`,
		"variables": map[string]any{
			"instrument_id": "BSVUSD",
			"limit":         800,
			"data_range": map[string]any{
				"time_from": from.UTC().Format(timeFmt),
				"time_to":   to.UTC().Format(timeFmt),
			},
			"periodicity": periodicity,
		},
	}
	return json.Marshal(m)
}

func getOgFrom(from time.Time, periodicity string) ([]TimeRate, error) {
	q, err := makeq(from, time.Now(), periodicity)
	if err != nil {
		return nil, err
	}

	res, err := http.Post(ogUrl, "application/json", bytes.NewReader(q))
	if err != nil {
		return nil, err
	}

	var js map[string]any
	err = json.NewDecoder(res.Body).Decode(&js)

	elements, ok := js["data"].(map[string]any)
	if !ok {
		return nil, errors.New("js[data] was expected to be a map")

	}

	dataPoints, ok := elements["instrument_price_bars"].([]any)
	if !ok {
		return nil, errors.New("elements[instrument_price_bars] expected to be a slice")
	}

	slices.Reverse(dataPoints)
	fmt.Println(len(dataPoints))
	rates := make([]TimeRate, 0, len(dataPoints))
	for _, x := range dataPoints {
		x_, ok := x.(map[string]any)
		if !ok {
			return nil, errors.New("expect x to be a map")
		}

		rate, ok := x_["close"].(float64)
		if !ok {
			return nil, errors.New("expect x_[close] to be a float64")
		}
		rate_ := float32(rate)

		timestr, ok := x_["ts"].(string)
		if !ok {
			return nil, errors.New("expect x_[ts] to be a string")
		}

		tt, err := time.Parse(minFmt, timestr)
		if err != nil {
			return nil, err
		}

		time := tt.Unix()
		rates = append(rates, TimeRate{Rate: rate_, Time: time})

	}

	return rates, nil
}

// func updateOGRates(path string) error {
// 	f, err := os.OpenFile(paths.HISTORICAL_RATES_FILE,
// 		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 	if err != nil {
// 		return err
// 	}
// 	defer f.Close()

// 	var from time.Time
// 	ts, err := latestTime(path)
// 	if err != nil {
// 		from = time.Now().Add(time.Hour * 24 * 2 * -1)
// 	} else {
// 		from = time.UnixMilli(ts * 100)
// 	}

// 	q, err := makeq(from, time.Now(), "minute15")
// 	if err != nil {
// 		return err
// 	}

// 	res, err := http.Post(ogateurl, "application/json", bytes.NewReader(q))
// 	if err != nil {
// 		return err
// 	}

// 	var js map[string]any
// 	err = json.NewDecoder(res.Body).Decode(&js)

// 	elements, ok := js["data"].(map[string]any)
// 	if !ok {
// 		return errors.New("js[data] was expected to be a map")

// 	}

// 	dataPoints, ok := elements["instrument_price_bars"].([]any)
// 	if !ok {
// 		return errors.New("elements[instrument_price_bars] was expected to be a slice")
// 	}

// 	slices.Reverse(dataPoints)
// 	fmt.Println(len(dataPoints))
// 	for _, x := range dataPoints {
// 		x_, ok := x.(map[string]any)
// 		if !ok {
// 			return errors.New("expect x to be a map")
// 		}

// 		rate, ok := x_["close"].(float64)
// 		if !ok {
// 			return errors.New("expect x_[close] to be a float64")
// 		}
// 		rate_ := float32(rate)

// 		timestr, ok := x_["ts"].(string)
// 		if !ok {
// 			return errors.New("expect x_[ts] to be a string")
// 		}

// 		tt, err := time.Parse(minFmt, timestr)
// 		if err != nil {
// 			return err
// 		}

// 		time := tt.UnixMilli() / 100
// 		err = binary.Write(f, binary.BigEndian, rate_)
// 		if err != nil {
// 			return err
// 		}

// 		err = binary.Write(f, binary.BigEndian, time)
// 		if err != nil {
// 			return err
// 		}

// 		fmt.Printf("time: %d, rate: %.2f\n", ts, rate_)
// 	}

// 	return nil

// }
