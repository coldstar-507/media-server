package paths

import (
	// "errors"
	// "log"
	// "os"
	"path/filepath"
)

var (
	_DOCKER_APP_PATH     = "/app/"
	_LOCAL_APP_PATH      = "/home/scott/dev/down4/go-custom-back/media-server/"
	APP_PATH             string
	DATA_PATH            string
	TEMPORARY_MEDIA_PATH string
	STATIC_MEDIA_PATH    string
	LOGS_PATH            string
)

func InitWD(testing bool) {
	if testing {
		APP_PATH = _LOCAL_APP_PATH
	} else {
		APP_PATH = _DOCKER_APP_PATH
	}
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

func MakeTempPath(id string) string {
	return MakeMediaPath(id, true)
}

func MakeStaticPath(id string) string {
	return MakeMediaPath(id, false)
}

func MakeMediaPath(id string, temporary bool) string {
	if temporary {
		return filepath.Join(TEMPORARY_MEDIA_PATH, id)
	} else {
		return filepath.Join(STATIC_MEDIA_PATH, id)
	}
}
