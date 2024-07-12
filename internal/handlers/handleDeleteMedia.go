package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/coldstar-507/media-server/internal/paths"
)

func RemoveMedia(id string, temp bool) error {
	path := paths.MakeMediaPath(id, temp)
	return os.Remove(path)
}

func HandleDeleteMedia(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := RemoveMedia(id, false); err != nil {
		log.Println("HandleDeleteMedia error removing file:", err)
		w.WriteHeader(500)
		return
	}
}
