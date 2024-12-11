module github.com/PonomarevAlexxander/queuing-system/messages

go 1.23.3

require (
	github.com/golang/protobuf v1.5.4
	google.golang.org/protobuf v1.35.2
)

replace (
	github.com/PonomarevAlexxander/queuing-system/messages => ../../generated/messages
	github.com/PonomarevAlexxander/queuing-system/services => ../../generated/services
	github.com/PonomarevAlexxander/queuing-system/utils => ../../utils
)

require github.com/google/go-cmp v0.6.0 // indirect
