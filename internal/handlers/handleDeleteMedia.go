package handlers

import (
	"net/http"
)

func HandleDeleteMedia(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	DeleteMedia(id)
}
