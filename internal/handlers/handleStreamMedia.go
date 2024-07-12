package handlers

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/coldstar-507/media-server/internal/paths"
)

func StreamMedia(id string, temp bool, w io.Writer) error {
	path := paths.MakeMediaPath(id, temp)
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("StreamMedia error opening file: %v", err)
	}
	defer f.Close()

	buf := make([]byte, 2)
	if _, err := f.Read(buf); err != nil {
		return fmt.Errorf("StreamMedia error reading metadata len: %v", err)
	}
	l := int64(binary.BigEndian.Uint16(buf))
	if _, err := f.Seek(l, io.SeekCurrent); err != nil {
		return fmt.Errorf("StreamMedia error seeking start of data: %v", err)
	}

	if _, err = io.Copy(w, f); err != nil && err != io.EOF {
		return fmt.Errorf("StreamMedia error copying file to writer: %v", err)
	}
	return nil
}

func HandleStreamMedia(w http.ResponseWriter, r *http.Request) {
	id, temp := r.PathValue("id"), r.PathValue("temp") == "true"
	if err := StreamMedia(id, temp, w); err != nil {
		w.WriteHeader(500)
		log.Println("HandleStreamMedia error: ", err)
	}
}
