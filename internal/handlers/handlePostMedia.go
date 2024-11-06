package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/coldstar-507/media-server/internal/paths"
)

func WriteMedia(id string, permanent bool, rdr io.Reader) error {
	path := paths.MakeMediaPath(id, permanent)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("WriteMedia error creating media file id=%s, %v", id, err)
	}
	defer f.Close()
	if _, err := io.Copy(f, rdr); err != nil && err != io.EOF {
		return fmt.Errorf("WriteMedia error saving media file id=%s: %v", id, err)
	}
	return nil
}

func WriteMedia2(id string, permanent bool, b []byte) error {
	path := paths.MakeMediaPath(id, permanent)
	return os.WriteFile(path, b, 0644)
}

func HandlePostMedia(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	permanent := paths.IsPermanent(id)
	defer r.Body.Close()
	if err := WriteMedia(id, permanent, r.Body); err != nil {
		log.Println("HandlePostMedia error: ", err)
		w.WriteHeader(500)
	}
}
