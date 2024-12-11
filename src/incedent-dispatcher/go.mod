module github.com/PonomarevAlexxander/queuing-system/incedent-dispatcher

go 1.23.3

replace (
	github.com/PonomarevAlexxander/queuing-system/messages => ../../generated/messages
	github.com/PonomarevAlexxander/queuing-system/services => ../../generated/services
	github.com/PonomarevAlexxander/queuing-system/utils => ../../utils
)

require (
	github.com/PonomarevAlexxander/queuing-system/messages v0.0.0-00010101000000-000000000000
	github.com/PonomarevAlexxander/queuing-system/services v0.0.0-00010101000000-000000000000
	github.com/PonomarevAlexxander/queuing-system/utils v0.0.0-00010101000000-000000000000
	github.com/alexflint/go-arg v1.5.1
	github.com/benbjohnson/clock v1.3.5
	go.uber.org/zap v1.27.0
	golang.org/x/sync v0.8.0
	google.golang.org/grpc v1.68.1
	google.golang.org/protobuf v1.35.2
)

require (
	github.com/alexflint/go-scalar v1.2.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.23.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/crypto v0.28.0 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	google.golang.org/genproto v0.0.0-20241206012308-a4fef0638583 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241118233622-e639e219e697 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
