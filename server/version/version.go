package version

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
)

func WithServerInterceptor() grpc.ServerOption {
	return grpc.UnaryInterceptor(serverInterceptor)
}

type Versioned interface {
	Version() string
}

func serverInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Check client's header
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		clientVersion := md.Get("api-version")
		log.Printf("Client API Version [%s]", clientVersion)
	}
	// Calls the handler
	result, err := handler(ctx, req)

	// Check if the server support the version
	if versioned := info.Server.(Versioned); versioned != nil {
		// Set server header
		log.Printf("Intercepted versioned, Server API Version [%s]", versioned.Version())
		header := metadata.Pairs("api-version", versioned.Version())
		if e := grpc.SendHeader(ctx, header); e != nil {
			log.Printf("Failed to add a header %s", e)
		}
	}
	return result, err
}

func WithClientInterceptor(version string, versionedMethods map[string]bool) grpc.DialOption {
	clientInterceptor := func(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Append client-side version
		if _, ok := versionedMethods[method]; ok {
			ctx = metadata.AppendToOutgoingContext(ctx, "api-version", version)
		}

		// TODO Add header extractor - for dev/log purpose only. Should be removed before release
		header := grpc.HeaderCallOption{}
		for _, o := range opts {
			if ho := o.(grpc.HeaderCallOption); ho.HeaderAddr != nil {
				header = ho
			}
		}
		if header.HeaderAddr == nil {
			md := metadata.New(make(map[string]string))
			header = grpc.Header(&md).(grpc.HeaderCallOption)
			opts = append(opts, header)
		}

		// Calls the invoker to execute RPC
		err := invoker(ctx, method, req, reply, cc, opts...)

		// Retrieve server-side version
		fmt.Println("Server Version: ", header.HeaderAddr.Get("api-version"))

		return err
	}
	return grpc.WithUnaryInterceptor(clientInterceptor)
}

