package main

import (
	"log"
	"net/http"

	"github.com/coldstar-507/media-server/internal/handlers"
	// "github.com/coldstar-507/media-server/internal/logger"
	"github.com/coldstar-507/media-server/internal/paths"
	"github.com/coldstar-507/utils"
)

func main() {
	// paths.CreateDirPaths()
	paths.InitWD(false)
	// logger.InitLogger() // We don't use this logger yet
	// defer logger.CloseLogger()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", handlers.HandlePing)

	mux.HandleFunc("GET /media/{id}/{temp}", handlers.HandleGetMedia)
	mux.HandleFunc("POST /media/{id}/{temp}", handlers.HandlePostMedia)
	mux.HandleFunc("DELETE /media/{id}", handlers.HandleDeleteMedia)

	mux.HandleFunc("GET /stream-media/{id}/{temp}", handlers.HandleStreamMedia)
	mux.HandleFunc("GET /metadata/{id}/{temp}", handlers.HandleGetMetadata)

	mux.HandleFunc("GET /payment/{id}", handlers.HandleGetPayment)
	mux.HandleFunc("POST /payment/{id}", handlers.HandlePostPayment)

	server := utils.ApplyMiddlewares(mux,
		utils.HttpLogging,
		utils.StatusLogger,
	)

	log.Println("Starting http media-server on 0.0.0.0:8081")
	if err := http.ListenAndServe("0.0.0.0:8081", server); err != nil {
		log.Fatalln("main error listening to http server:", err)
	}
}
