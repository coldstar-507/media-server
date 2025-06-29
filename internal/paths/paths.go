package paths

import (
	"errors"
	"os"
	"path"
	"path/filepath"

	"github.com/coldstar-507/utils2"
)

var (
	ServerFolder           string
	TEMPORARY_PATH         string
	PERMANENT_PATH         string
	LOGS_PATH              string
	RATES_PATH             string
	HISTORICAL_RATES_FILE  string
	FIFTEEN_MIN_RATES_FILE string
	ONE_HOUR_RATES_FILE    string
	TEMP_IDS_FILE          string
	PERM_IDS_FILE          string
)

func InitWD() {
	ServerFolder = os.Getenv("SERVER_FOLDER")
	utils2.Assert(len(ServerFolder) > 0, "Undefined SERVER_FOLDER")
	PERMANENT_PATH = path.Join(ServerFolder, "permanent")
	TEMPORARY_PATH = path.Join(ServerFolder, "temporary")
	errMkdirs := errors.Join(
		os.MkdirAll(path.Join(PERMANENT_PATH, "meta"), 0755),
		os.MkdirAll(path.Join(PERMANENT_PATH, "data"), 0755),
		os.MkdirAll(path.Join(PERMANENT_PATH, "thum"), 0755),
		os.MkdirAll(path.Join(PERMANENT_PATH, "temp"), 0755),

		os.MkdirAll(path.Join(TEMPORARY_PATH, "meta"), 0755),
		os.MkdirAll(path.Join(TEMPORARY_PATH, "data"), 0755),
		os.MkdirAll(path.Join(TEMPORARY_PATH, "thum"), 0755),
		os.MkdirAll(path.Join(TEMPORARY_PATH, "temp"), 0755),
	)
	utils2.Must(errMkdirs)
	// PERMANENT_PATH = os.Getenv("PERMANENT_PATH")
	// TEMPORARY_PATH = os.Getenv("TEMPORARY_PATH")
	LOGS_PATH = os.Getenv("LOGS_PATH")
	RATES_PATH = os.Getenv("RATES_PATH")

	// if len(PERMANENT_PATH) == 0 {
	// 	panic("PERMANENT_PATH undefined, please check ENV file at root of project")
	// } else if len(TEMPORARY_PATH) == 0 {
	// 	panic("TEMPORARY_PATH undefined, please check ENV file at root of project")
	// } else if len(LOGS_PATH) == 0 {
	// 	panic("LOGS_PATH undefined, please check ENV file at root of project")
	// } else if len(RATES_PATH) == 0 {
	// 	panic("RATES_PATH undefined, please check ENV file at root of project")
	// }

	TEMP_IDS_FILE = filepath.Join(TEMPORARY_PATH, "ids")
	PERM_IDS_FILE = filepath.Join(PERMANENT_PATH, "ids")
	HISTORICAL_RATES_FILE = filepath.Join(RATES_PATH, "historical")
	ONE_HOUR_RATES_FILE = filepath.Join(RATES_PATH, "one_hour_rates")
	FIFTEEN_MIN_RATES_FILE = filepath.Join(RATES_PATH, "fifteen_min_rates")
}

func IsPermanent(id string) bool {
	return id[len(id)-2:] == "01"
}

// func MakeTempPath(id string) string {
// 	return MakeMediaPath(id, false)
// }

// func MakeStaticPath(id string) string {
// 	return MakeMediaPath(id, true)
// }

func MakePath(folder, id string) string {
	if IsPermanent(id) {
		return filepath.Join(PERMANENT_PATH, folder, id)
	} else {
		return filepath.Join(TEMPORARY_PATH, folder, id)
	}
}

func MakePathExt(folder, id, ext string) string {
	if len(ext) == 0 {
		return MakePath(folder, id)
	}
	if IsPermanent(id) {
		return filepath.Join(PERMANENT_PATH, folder, id+"."+ext)
	} else {
		return filepath.Join(TEMPORARY_PATH, folder, id+"."+ext)
	}
}

func MakeMediaPath(id string, permanent bool) string {
	if permanent {
		return filepath.Join(PERMANENT_PATH, id)
	} else {
		return filepath.Join(TEMPORARY_PATH, id)

	}
}
