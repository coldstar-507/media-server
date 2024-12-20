package main

import (
	"log"
	"net/http"

	"github.com/coldstar-507/media-server/internal/handlers"
	"github.com/coldstar-507/media-server/internal/paths"
	"github.com/coldstar-507/router/router_utils"
	"github.com/coldstar-507/utils/http_utils"
)

// files are saved on disk in this way:
// as a single []byte
// {media data} is a common media, like jpg, png, mp4, etc..
// [u16 len of metadata, metadata flatbuffer, media data]

// this should be part of the ENV in production
var (
	ip         string                     = "localhost"
	place      router_utils.SERVER_NUMBER = "0x0100"
	routerType router_utils.ROUTER_TYPE   = router_utils.MEDIA_ROUTER
)

func main() {

	log.Println("Starting hook manager")
	go handlers.HookManager.Run()

	log.Println("Starting local router")
	router_utils.InitLocalServer(ip, place, routerType)
	go router_utils.LocalServer.Run()

	paths.InitWD()

	// logger.InitLogger() // We don't use this logger yet
	// defer logger.CloseLogger()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", router_utils.HandlePing)
	mux.HandleFunc("GET /route-scores", router_utils.HandleScoreRequest)
	mux.HandleFunc("GET /local-router", router_utils.HandleServerStatus)
	mux.HandleFunc("GET /full-router", router_utils.HandleRouterStatus)

	mux.HandleFunc("GET /media/{id}", handlers.HandleGetMedia)
	mux.HandleFunc("POST /media/{id}", handlers.HandlePostMedia)
	mux.HandleFunc("DELETE /media/{id}", handlers.HandleDeleteMedia)

	mux.HandleFunc("GET /stream-media/{id}", handlers.HandleStreamMedia)
	mux.HandleFunc("GET /metadata/{id}", handlers.HandleGetMetadata)

	// payments are currently sent to chat server with pushId
	// they aren't though
	mux.HandleFunc("GET /payment/{id}", handlers.HandleGetPayment)
	mux.HandleFunc("POST /payment/{id}", handlers.HandlePostPayment)

	mux.HandleFunc("POST /generate-media", handlers.HandleGenerateMedia)
	mux.HandleFunc("POST /generate-media-hook", handlers.HandleGenerateMediaHook)

	server := http_utils.ApplyMiddlewares(mux, http_utils.StatusLogger)

	addr := "0.0.0.0:8081"
	// crt, key := "../service-accounts/server.crt", "../service-accounts/server.key"
	// err := http.ListenAndServeTLS(addr, crt, key, server)
	log.Println("media-server listening on", addr)
	err := http.ListenAndServe(addr, server)
	log.Println("stoping media-server:", err)
}
