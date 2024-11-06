package main

import (
	"github.com/coldstar-507/media-server/internal/handlers"
)

func main() {
	uid := handlers.Uuid()
	print(uid)
}
