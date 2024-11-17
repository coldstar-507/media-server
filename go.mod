module github.com/coldstar-507/media-server

go 1.23.2

require (
	github.com/coldstar-507/router v0.0.0
	github.com/coldstar-507/utils/http_utils v0.0.0
	github.com/vmihailenco/msgpack/v5 v5.4.1
)

replace (
	github.com/coldstar-507/flatgen => ../../flatbufs/flatgen
	github.com/coldstar-507/router => ../router-server
	github.com/coldstar-507/utils/http_utils => ../utils/http_utils
)

require (
	github.com/stretchr/testify v1.8.4 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
)
