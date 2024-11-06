package paths

import (
	// "errors"
	// "log"
	// "os"
	// "log"
	// "os"
	"path/filepath"
	// "github.com/coldstar-507/utils"
)

var (
	_DOCKER_APP_PATH     = "/app/"
	_LOCAL_APP_PATH      = "./"
	APP_PATH             string
	DATA_PATH            string
	TEMPORARY_MEDIA_PATH string
	STATIC_MEDIA_PATH    string
	LOGS_PATH            string
)

func InitWD(appPath string) {
	APP_PATH = appPath
	DATA_PATH = filepath.Join(APP_PATH, "data")
	TEMPORARY_MEDIA_PATH = filepath.Join(DATA_PATH, "temp")
	STATIC_MEDIA_PATH = filepath.Join(DATA_PATH, "static")
	LOGS_PATH = filepath.Join(DATA_PATH, "logs")
}

// func CreateDirPaths() {
// 	err := errors.Join(
// 		os.MkdirAll(DATA_PATH, 0755),
// 		os.MkdirAll(TEMPORARY_MEDIA_PATH, 0755),
// 		os.MkdirAll(STATIC_MEDIA_PATH, 0755),
// 		os.MkdirAll(LOGS_PATH, 0755))
// 	if err != nil {
// 		log.Fatalln("CreateDirPaths error: ", err)
// 	}
// }

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
		return filepath.Join(STATIC_MEDIA_PATH, id)
	} else {
		return filepath.Join(TEMPORARY_MEDIA_PATH, id)
	}
}
