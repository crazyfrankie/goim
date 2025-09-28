#!/bin/bash

# 生成所有proto文件
protoc \
  --proto_path=idl \
  --go_out=. \
  --go_opt=module=github.com/crazyfrankie/goim \
  --go-grpc_out=. \
  --go-grpc_opt=module=github.com/crazyfrankie/goim \
  idl/common/v1/common.proto \
  idl/user/v1/user.proto

echo "Proto generation completed successfully"
