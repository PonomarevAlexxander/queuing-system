.PHONY: run-ut
run-ut:
	ginkgo ./...

.PHONY: run-generate
run-generate:
	ginkgo ./...

.PHONY: build
build:
	go build -o out/incedent-dispatcher src/incedent-dispatcher/cmd/main.go &&\
	go build -o out/incedent-processing-service src/incedent-processing-service/cmd/main.go &&\
	go build -o out/incedent-producer-service src/incedent-producer-service/cmd/main.go

.PHONY: emulate
emulate:
	./out/incedent-producer-service --config src/incedent-producer-service/config/config.yaml --priority 1 \
  ./out/incedent-dispatcher --config src/incedent-dispatcher/config/config.yaml \
  ./out/incedent-processing-service --id 1 --host localhost:8090 --config src/incedent-processing-service/config/config.yaml

.PHONY: go-get
go-get:
	cd utils && go get ./... && cd ../ && \
  cd src/incedent-dispatcher && go get ./... && cd ../../ && \
  cd src/incedent-processing-service && go get ./... && cd ../../ && \
  cd generated/messages && go get ./... && cd ../../ && \
  cd generated/services && go get ./... && cd ../../ && \
  cd src/incedent-producer-service && go get ./... && cd ../../

.PHONY: go-tidy
go-tidy:
	cd utils && go mod tidy && cd ../ && \
  cd src/incedent-dispatcher && go mod tidy && cd ../../ && \
  cd src/incedent-processing-service && go mod tidy && cd ../../ && \
  cd generated/messages && go mod tidy && cd ../../ && \
  cd generated/services && go mod tidy && cd ../../ && \
  cd src/incedent-producer-service && go mod tidy && cd ../../

.PHONY: proto-generate
proto-generate:
	protoc --proto_path=protos --go_out=generated --go_opt=module=github.com/PonomarevAlexxander/queuing-system \
	--go-grpc_out=generated --go-grpc_opt=module=github.com/PonomarevAlexxander/queuing-system \
	messages/common/types.proto messages/incedent/incedent.proto messages/registration/registration.proto \
	services/incedent_dispatcher/incedent_dispatcher.proto \
	services/incedent_processor/incedent_processor.proto
