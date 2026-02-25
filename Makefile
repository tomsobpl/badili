# Path to your proto files
PROTO_DIR = api/gelfapi/v1

.PHONY: proto
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/*.proto

.PHONY: clean
clean:
	rm -f $(PROTO_DIR)/*.pb.go

.PHONY: docker-build
docker-build:
	docker build -t badili:latest .

.PHONY: docker-clean
docker-clean:
	docker rmi badili:latest
