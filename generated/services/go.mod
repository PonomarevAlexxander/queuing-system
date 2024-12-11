module github.com/PonomarevAlexxander/queuing-system/services

go 1.23.3

replace (
	github.com/PonomarevAlexxander/queuing-system/messages => ../../generated/messages
	github.com/PonomarevAlexxander/queuing-system/utils => ../../utils
)

require (
	github.com/PonomarevAlexxander/queuing-system/messages v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.68.0
	google.golang.org/protobuf v1.35.2
)

require (
	github.com/golang/protobuf v1.5.4 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	google.golang.org/genproto v0.0.0-20241206012308-a4fef0638583 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241118233622-e639e219e697 // indirect
)
