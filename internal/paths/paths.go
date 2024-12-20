package paths

import (
	"os"
	"path/filepath"
)

var (
	TEMPORARY_PATH string
	PERMANENT_PATH string
	LOGS_PATH      string
)

func InitWD() {
	PERMANENT_PATH = os.Getenv("PERMANENT_PATH")
	TEMPORARY_PATH = os.Getenv("TEMPORARY_PATH")
	LOGS_PATH = os.Getenv("LOGS_PATH")

	if len(PERMANENT_PATH) == 0 {
		panic("PERMANENT_PATH undefined, please check ENV file at root of project")
	} else if len(TEMPORARY_PATH) == 0 {
		panic("TEMPORARY_PATH undefined, please check ENV file at root of project")
	} else if len(LOGS_PATH) == 0 {
		panic("LOGS_PATH undefined, please check ENV file at root of project")
	}
}

func IsPermanent(id string) bool {
	return id[len(id)-2:] == "01"
}

func MakeTempPath(id string) string {
	return MakeMediaPath(id, false)
}

func MakeStaticPath(id string) string {
	return MakeMediaPath(id, true)
}

func MakeMediaPath(id string, permanent bool) string {
	if permanent {
		return filepath.Join(PERMANENT_PATH, id)
	} else {
		return filepath.Join(TEMPORARY_PATH, id)

	}
}
