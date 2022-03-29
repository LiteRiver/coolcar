#!/bin/bash

function genProto {
  DOMAIN=$1
  PROTO_PATH=./${DOMAIN}/api
  GO_OUT_PATH=./${DOMAIN}/api/gen/v1
  mkdir -p $GO_OUT_PATH

  protoc -I $PROTO_PATH \
    --go_out $GO_OUT_PATH --go_opt paths=source_relative \
    --go-grpc_out $GO_OUT_PATH --go-grpc_opt paths=source_relative \
    $PROTO_PATH/${DOMAIN}.proto

  protoc -I $PROTO_PATH \
    --grpc-gateway_out $GO_OUT_PATH \
    --grpc-gateway_opt logtostderr=true \
    --grpc-gateway_opt paths=source_relative \
    --grpc-gateway_opt grpc_api_configuration=$PROTO_PATH/${DOMAIN}.yaml \
    $PROTO_PATH/${DOMAIN}.proto

  PBJS_BIN=../wx/miniprogram/node_modules/.bin
  PBJS_OUT=../wx/miniprogram/services/proto-gen/${DOMAIN}
  mkdir -p $PBJS_OUT
  printf "import * as \$protobuf from \"protobufjs\";\n" >$PBJS_OUT/${DOMAIN}-pb.js
  $PBJS_BIN/pbjs -t static -w es6 $PROTO_PATH/${DOMAIN}.proto \
    --no-create --no-encode --no-decode --no-verify \
    --no-delimited --force-number >>$PBJS_OUT/${DOMAIN}-pb.js
  $PBJS_BIN/pbts -o $PBJS_OUT/${DOMAIN}-pb.d.ts $PBJS_OUT/${DOMAIN}-pb.js
}

genProto auth
genProto rental
