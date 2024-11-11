module github.com/coldstar-507/media-server

go 1.23.2

require (
	github.com/coldstar-507/utils v0.0.0-20241106185519-845eee7ad9d5
	github.com/vmihailenco/msgpack/v5 v5.4.1
)

replace (
	github.com/coldstar-507/flatgen => ../../flatbufs/flatgen
	github.com/coldstar-507/utils => ../utils
)

require (
	github.com/btcsuite/btcd/btcutil v1.1.6 // indirect
	github.com/coldstar-507/flatgen v0.0.0-20240830172816-703a5c6098f5 // indirect
	github.com/google/flatbuffers v24.3.25+incompatible // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	go.mongodb.org/mongo-driver v1.17.1 // indirect
)
