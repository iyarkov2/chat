##GO Reflection Experiment

There are 3 different go modules
1. **api** module contains nothing but proto files. There are chat proto file and google's timestamp file.
The only reason I added google proto file is that I experimented a bit with better way to deal with proto dependencies  
2. **server** uses go files compiled by _protoc_. There are two main files:
   1. **marshal/main.go** produces _out/message.out_ file. It is a binary file contains a single ConnectRequest object.
   1. **server/main.go** is a gRPC server
3. **client** is go gRPC client that uses protobuf registry and protobuf reflection to decode
the binary file and make gRPC calls