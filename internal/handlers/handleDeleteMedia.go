package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/coldstar-507/media-server/internal/paths"
)

func RemoveMedia(id string, isPermanent bool) error {
	path := paths.MakeMediaPath(id, isPermanent)
	return os.Remove(path)
}

func HandleDeleteMedia(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	isPermanent := paths.IsPermanent(id)
	if err := RemoveMedia(id, isPermanent); err != nil {
		log.Println("HandleDeleteMedia error removing file:", err)
		w.WriteHeader(500)
		return
	}
}
