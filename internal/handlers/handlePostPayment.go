package handlers

import (
	"log"
	"net/http"
)

func HandlePostPayment(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	id := r.PathValue("id")
	if err := WriteMedia(id, false, r.Body); err != nil {
		log.Println("HandlePostPayment error:", err)
		w.WriteHeader(500)
	}
}
