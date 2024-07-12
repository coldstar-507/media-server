package logger

import (
	"log"
	"os"
	"path/filepath"

	"github.com/coldstar-507/media-server/internal/paths"
)

var (
	logFile *os.File
	err     error
)

func InitLogger() {
	wd, er := os.Getwd()
	if er != nil {
		log.Fatalln("Getwd error: ", er)
	}
	log.Println("Wd: ", wd)
	path := filepath.Join(paths.LOGS_PATH, "logs.txt")
	flags := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	if logFile, err = os.OpenFile(path, flags, 0644); err != nil {
		log.Fatalln("InitLogger error opening log file: ", err)
	} else {
		log.SetOutput(logFile)
	}
}

func CloseLogger() {
	logFile.Close()
}
