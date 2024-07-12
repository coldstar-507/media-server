package main

import (
	// "fmt"
	"log"
	// "math"
	"os"
	// "strconv"
	// "time"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("CreateDirPaths error on os.Getwd:", err)
	}

	log.Println("wd=", wd)

	// t1 := time.Now()

	// fmt.Println(t1.UnixMilli())
	// fmt.Println(math.MaxUint32)
	// mu := strconv.FormatUint(math.MaxUint64, 10)
	// fmt.Println(mu)

}
