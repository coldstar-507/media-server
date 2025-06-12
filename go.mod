module github.com/coldstar-507/media-server

go 1.23.2

require (
	github.com/coldstar-507/flatgen v0.0.0-20240830172816-703a5c6098f5
	github.com/coldstar-507/router v0.0.0
	github.com/coldstar-507/utils/http_utils v0.0.0
	github.com/coldstar-507/utils/id_utils v0.0.0
	github.com/coldstar-507/utils/utils v0.0.0
	golang.org/x/sync v0.8.0
)

replace (
	github.com/coldstar-507/flatgen => ../../flatbufs/flatgen
	github.com/coldstar-507/router => ../router-server
	github.com/coldstar-507/utils/http_utils => ../utils/http_utils
	github.com/coldstar-507/utils/id_utils => ../utils/id_utils
	github.com/coldstar-507/utils/utils => ../utils/utils
)

require (
	github.com/google/flatbuffers v25.2.10+incompatible // indirect
	go.mongodb.org/mongo-driver v1.17.3 // indirect
)
