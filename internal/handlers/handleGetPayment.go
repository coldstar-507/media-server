package handlers

import (
	"log"
	"net/http"
)

func HandleGetPayment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := ReadMedia(id, true, w); err != nil {
		log.Println("HandleGetPayment error: ", err)
		w.WriteHeader(500)
	}
}
