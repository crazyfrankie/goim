.PHONY: proto clean-proto buf-gen

errcode:
	@echo "Generating error code"
	@./scripts/gen-error.sh --biz "*"

proto:
	@echo "Generating protobuf code with protoc..."
	@./scripts/gen-proto.sh

buf-gen:
	@echo "Generating protobuf code with buf..."
	@buf generate

clean-proto:
	@echo "Cleaning generated protobuf code..."
	@rm -rf protocol/

regen-proto: clean-proto proto
