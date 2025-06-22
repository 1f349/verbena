PROTOC := $(shell which protoc)

PROTO_SRC_DIR := internal/proto/src
PROTO_OUT_DIR := internal/proto/gen
PROTO_FILES := $(wildcard $(PROTO_SRC_DIR)/*.proto)

.PHONY: all protobuf

all: protobuf

protobuf: $(PROTO_FILES)
	@echo "Generating Go code from .proto files..."
	mkdir -p $(PROTO_OUT_DIR)
	rm -f $(PROTO_OUT_DIR)/*.pb.go
	$(PROTOC) -I $(PROTO_SRC_DIR) --go_out=$(PROTO_OUT_DIR) --go-grpc_out=$(PROTO_OUT_DIR) $(PROTO_FILES)
	@echo "Generation complete!"
