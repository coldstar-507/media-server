module github.com/coldstar-507/media-server

go 1.22.1

require (
	github.com/coldstar-507/utils v0.0.0-20240628170819-894d2f147162
	github.com/vmihailenco/msgpack/v5 v5.4.1
)

replace (
	github.com/coldstar-507/flatgen => ../../flatbufs/flatgen
	github.com/coldstar-507/utils => ../utils
)

require (
	github.com/btcsuite/btcd/btcutil v1.1.5 // indirect
	github.com/coldstar-507/flatgen v0.0.0-20240721154545-7f7a3c686f6f // indirect
	github.com/google/flatbuffers v24.3.25+incompatible // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	go.mongodb.org/mongo-driver v1.16.1 // indirect
)
