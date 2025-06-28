package main

import (
	"log"
	"net/http"

	"github.com/coldstar-507/media-server/internal/config"
	"github.com/coldstar-507/media-server/internal/handlers"
	"github.com/coldstar-507/media-server/internal/paths"
	"github.com/coldstar-507/router-server/router_utils"
	"github.com/coldstar-507/utils2"
)

func main() {

	log.Println("Starting hook manager")
	// go handlers.HookManager.Run()

	log.Println("Starting local router")
	conf := config.LoadConfig()
	router_utils.InitLocalServer(conf.SERVER_IP, conf.SERVER_PLACE, conf.SERVER_TYPE)
	log.Printf("SERVER_IP=%s, SERVER_PLACE=%d, SERVER_TYPE=%s\n",
		conf.SERVER_IP, conf.SERVER_PLACE, conf.SERVER_TYPE)

	go router_utils.LocalServer.Run()

	paths.InitWD()
	go handlers.PriceAgent.Run()

	go handlers.RunMediaWriteRequestsHandler()

	handlers.InitCoords()

	// logger.InitLogger() // We don't use this logger yet
	// defer logger.CloseLogger()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", router_utils.HandlePing)
	mux.HandleFunc("GET /route-scores", router_utils.HandleScoreRequest)
	mux.HandleFunc("GET /local-router", router_utils.HandleServerStatus)
	mux.HandleFunc("GET /full-router", router_utils.HandleRouterStatus)

	mux.HandleFunc("GET /storage-info", handlers.HandleGetStorageInfo)

	mux.HandleFunc("GET /media/{id}", handlers.HandleGetMedia)
	mux.HandleFunc("POST /media/{thum}", handlers.HandlePostMedia)
	mux.HandleFunc("POST /media", handlers.HandlePostMedia)
	mux.HandleFunc("DELETE /media/{id}", handlers.HandleDeleteMedia)
	mux.HandleFunc("GET /metadata/{id}", handlers.HandleGetMetadata)
	mux.HandleFunc("GET /thumbnail/{id}", handlers.HandleGetThumbnail)

	// mux.HandleFunc("GET /stream-media/{id}", handlers.HandleStreamMedia)

	mux.HandleFunc("GET /rates/{until}/{periodicity}", handlers.HandleRatesRequest)
	// payments are currently sent to chat server with pushId
	// they aren't though
	// mux.HandleFunc("GET /payment/{id}", handlers.HandleGetPayment)
	// mux.HandleFunc("POST /payment/{id}", handlers.HandlePostPayment)

	// mux.HandleFunc("POST /generate-media", handlers.HandleGenerateMedia)
	// mux.HandleFunc("POST /generate-media-hook", handlers.HandleGenerateMediaHook)

	mux.HandleFunc("GET /country-coords/{country}", handlers.HandleGetCoords)
	mux.HandleFunc("GET /countries", handlers.HandleGetCountries)
	mux.HandleFunc("GET /countries-pretty", handlers.HandleGetCountriesPretty)

	server := utils2.ApplyMiddlewares(mux, utils2.StatusLogger)

	addr := "0.0.0.0:8081"
	// crt, key := "../service-accounts/server.crt", "../service-accounts/server.key"
	// err := http.ListenAndServeTLS(addr, crt, key, server)
	log.Println("media-server listening on", addr)
	err := http.ListenAndServe(addr, server)
	log.Println("stoping media-server:", err)
}
