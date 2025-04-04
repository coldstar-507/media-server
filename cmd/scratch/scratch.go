package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var (
	xurl = "https://api.whatsonchain.com/v1/bsv/main/exchangerate/historical?from=%d&to=%d"
)

func makeUrl(from, to int64) string {
	return fmt.Sprintf(xurl, from, to)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func appendFrom(latest int64) error {
	now := time.Now().Unix()
	f, err := os.OpenFile("./rates", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}

	url := makeUrl(latest, now)
	fmt.Println(url)
	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error making http request: %w", err)
	}

	var j []struct {
		Rate float32 `json:"rate"`
		Time int64   `json:"time"`
	}

	err = json.NewDecoder(res.Body).Decode(&j)
	if err != nil {
		return fmt.Errorf("error decoding json response: %w", err)
	}

	for _, m := range j {
		if latest == m.Time {
			fmt.Printf("skipping %d (equals latest)\n", m.Time)
			continue
		}
		fmt.Printf("rate=%f, time=%d\n", m.Rate, m.Time)
		binary.Write(f, binary.BigEndian, m.Rate)
		binary.Write(f, binary.BigEndian, m.Time)
	}

	err = f.Close()
	if err != nil {
		return fmt.Errorf("error closing file: %w", err)
	}
	return nil
}

func write() {
	now := time.Now()
	weekAgo := now.Add(time.Hour * 24 * 2 * -1)

	f, err := os.Create("./rates")
	must(err)

	url := makeUrl(weekAgo.Unix(), now.Unix())
	fmt.Println(url)
	res, err := http.Get(url)
	must(err)

	var j []struct {
		Rate float32 `json:"rate"`
		Time int64   `json:"time"`
	}

	err = json.NewDecoder(res.Body).Decode(&j)
	must(err)

	for _, m := range j {
		fmt.Printf("rate=%f, time=%d\n", m.Rate, m.Time)
		binary.Write(f, binary.BigEndian, m.Rate)
		binary.Write(f, binary.BigEndian, m.Time)
	}

	err = f.Close()
	must(err)
	// must(err)
}

func main() {
	now := time.Now()
	weekAgo := now.Add(-1 * time.Hour * 24 * 7)
	fmt.Println(appendFrom(weekAgo.Unix()))
	lt, err := latestTime()
	fmt.Println(err)
	fmt.Println(appendFrom(lt))
	buf := bytes.NewBuffer(make([]byte, 0, 512))
	fmt.Println(readUntil(buf, weekAgo.Unix()))
}

func readUntil(w io.Writer, until int64) error {
	// Open the file
	file, err := os.Open("./rates")
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
	var time int64
	var rate float32
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

func latestTime() (int64, error) {
	file, err := os.Open("./rates")
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return 0, fmt.Errorf("failed to stat file: %w", err)
	}

	fileSize := fileInfo.Size()

	blockSize := int64(8)
	position := fileSize - blockSize

	_, err = file.Seek(position, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to seek: %w", err)
	}

	var time int64
	if err = binary.Read(file, binary.BigEndian, &time); err != nil {
		return 0, fmt.Errorf("failed to read time: %w", err)
	}
	return time, nil
}
