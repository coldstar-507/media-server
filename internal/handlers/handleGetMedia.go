package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/coldstar-507/media-server/internal/paths"
)

func ReadMedia(id string, permanent bool, w io.Writer) error {
	path := paths.MakeMediaPath(id, permanent)
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("ReadMedia error opening file: %v", err)
	}
	defer f.Close()
	if _, err := io.Copy(w, f); err != nil && err != io.EOF {
		return fmt.Errorf("ReadMedia error copying file: %v", err)
	}
	return nil
}

func HandleGetMedia(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	permanent := paths.IsPermanent(id)
	if err := ReadMedia(id, permanent, w); err != nil {
		log.Printf("HandleGetMedia error id=%s, permanent=%v: %v\n", id, permanent, err)
		w.WriteHeader(500)
	}
}
