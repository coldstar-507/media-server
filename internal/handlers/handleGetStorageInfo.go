package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	media_utils "github.com/coldstar-507/media-server/internal/utils"
)

func HandleGetStorageInfo(w http.ResponseWriter, r *http.Request) {
	total, available, err := media_utils.GetStorageInfo()
	if err != nil {
		log.Println("HandleGetStorageInfo: error getting storage info:", err)
		w.WriteHeader(500)
	} else {
		data := map[string]any{
			"total":     total,
			"available": available,
		}
		err = json.NewEncoder(w).Encode(data)
		if err != nil {
			log.Println("HandleGetStorageInfo: error encoding json response:", err)
			w.WriteHeader(501)
		}
	}
}
