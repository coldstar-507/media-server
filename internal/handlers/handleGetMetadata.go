package handlers

// import (
// 	"encoding/binary"
// 	"fmt"
// 	"io"
// 	"log"
// 	"net/http"
// 	"os"

// 	"github.com/coldstar-507/media-server/internal/paths"
// )

// func ReadMetadata(id string, permanent bool, w io.Writer) error {
// 	path := paths.MakeMediaPath(id, permanent)
// 	f, err := os.Open(path)
// 	if err != nil {
// 		return fmt.Errorf("ReadMetadata error opening file: %v", err)
// 	}
// 	defer f.Close()
// 	buf := make([]byte, 2)
// 	if _, err := f.Read(buf); err != nil {
// 		return fmt.Errorf("ReadMetadata error reading metadata len: %v", err)
// 	}

// 	l := int64(binary.BigEndian.Uint16(buf))
// 	if _, err := io.CopyN(w, f, l); err != nil {
// 		return fmt.Errorf("ReadMetadata error copying file: %v", err)
// 	}
// 	return nil
// }

// func HandleGetMetadata(w http.ResponseWriter, r *http.Request) {
// 	id := r.PathValue("id")
// 	permanent := paths.IsPermanent(id)
// 	if err := ReadMetadata(id, permanent, w); err != nil {
// 		log.Println("HandleGetMetadata error: ", err)
// 		w.WriteHeader(500)
// 	}
// }
