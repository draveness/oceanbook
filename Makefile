.PHONY: test

pb:
	protoc --go_out=plugins=grpc:$(GOPATH)/src -I api/protobuf-spec/oceanbookpb -I $(GOPATH)/src api/protobuf-spec/**/*.proto

clean:
	rm -rf tmp
	mkdir tmp

bin/oceanbook: build

build:
	go build -o bin/oceanbook cmd/oceanbook/*.go

start: build
	goreman start

test:
	go test ./...
