#tools:
#	go install google.golang.org/protobuf/cmd/protoc-gen-go @v1.27.1

download:
	go mod tidy
	go mod download

generate:
	protoc --proto_path=/usr/local/Cellar/protobuf/3.17.3/include --proto_path=. --go_out=paths=source_relative:. chat.proto
