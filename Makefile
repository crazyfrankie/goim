.PHONY: proto clean-proto buf-gen buf-lint

errcode:
	@echo "Generating error code"
	@./scripts/gen-error.sh --biz "*"

proto:
	@echo "Generating protobuf code with protoc..."
	@./scripts/gen-proto.sh

buf-gen:
	@echo "Generating protobuf code with buf..."
	@buf generate --template scripts/buf/buf.gen.yaml

buf-lint:
	@echo "Linting protobuf files with buf..."
	@buf lint --config scripts/buf/buf.yaml

clean-proto:
	@echo "Cleaning generated protobuf code..."
	@rm -rf protocol/

regen-proto: clean-proto proto
