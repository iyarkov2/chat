#!/bin/bash

rm -f api/*

echo 'Generating go API files from protobuffer files'

protoc -I $PROTOBUF_INCLUDE --proto_path=../api --go_out=api --go_opt=paths=source_relative --go_opt=Mchat.proto=github.com/iyarkov2/chat/client/api chat.proto
protoc -I $PROTOBUF_INCLUDE --proto_path=../api --go-grpc_out=api --go-grpc_opt=paths=source_relative --go-grpc_opt=Mchat.proto=github.com/iyarkov2/chat/client/api chat.proto
protoc -I $PROTOBUF_INCLUDE --proto_path=../api --go-version_out=api --go-version_opt=paths=source_relative --go-version_opt=rev=dc0a94c --go-version_opt=Mchat.proto=github.com/iyarkov2/chat/client/api chat.proto

#
#  --go-grpc_out=api --go-grpc_opt=paths=source_relative
#echo 'generating mocks'
# go run github.com/golang/mock/mockgen -package $GOPACKAGE -source=payment_agreement_service.pb.go -destination=mock_payment_agreement_server.go -self_package=. PaymentAgreementServiceServer

