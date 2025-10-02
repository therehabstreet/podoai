PROTO_DIR=proto
PROTO_OUT=proto
PROTOC=protoc

# List all proto files recursively
PROTO_FILES=$(shell find $(PROTO_DIR) -name '*.proto')

.PHONY: gen clean

gen:
	@echo "Generating Go code from proto files..."
	$(PROTOC) \
		--go_out=$(PROTO_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT) --go-grpc_opt=paths=source_relative \
		-I $(PROTO_DIR) \
		$(PROTO_FILES)

clean:
	@echo "Cleaning generated files..."
	find $(PROTO_OUT) -name '*.pb.go' -delete
