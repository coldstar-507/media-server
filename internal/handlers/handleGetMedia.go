package handlers

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/coldstar-507/media-server/internal/paths"
	// media_utils "github.com/coldstar-507/media-server/internal/utils"
)

func ReadFileTo(w io.Writer, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(w, f); err != nil && err != io.EOF {
		return err
	}
	return nil
}

func HandleGetMedia(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	path := paths.MakePathExt("data", id, "webp")
	if err := ReadFileTo(w, path); err != nil {
		log.Printf("HandleGetMedia error path=%s: %v\n", path, err)
		w.WriteHeader(500)
	}
}

func HandleGetThumbnail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	path := paths.MakePathExt("thum", id, "webp")
	if err := ReadFileTo(w, path); err != nil {
		log.Printf("HandleGetThumbnail error path=%s: %v\n", path, err)
		w.WriteHeader(500)
	}
}

func HandleGetMetadata(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	path := paths.MakePath("meta", id)
	if err := ReadFileTo(w, path); err != nil {
		log.Printf("HandleGetMetadata error path=%s: %v\n", id, err)
		w.WriteHeader(500)
	}
}
