module github.com/iyarkov2/chat/server

go 1.17

require (
	github.com/iyarkov2/chat/api v0.0.0
	github.com/rs/zerolog v1.24.0
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
)

require (
	github.com/golang/protobuf v1.5.0 // indirect
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4 // indirect
	golang.org/x/sys v0.0.0-20210510120138-977fb7262007 // indirect
	golang.org/x/text v0.3.3 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
)

replace github.com/iyarkov2/chat/api v0.0.0 => ../api
