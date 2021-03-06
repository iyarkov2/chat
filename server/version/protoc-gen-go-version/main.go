package main

import (
	"flag"
	"fmt"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/pluginpb"
)

var rev *string

func main() {
	var flags flag.FlagSet
	rev = flags.String("rev", "0", "Git Revision")

	protogen.Options{ ParamFunc: flags.Set }.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		// The type information for all extensions is in the source files,
		// so we need to extract them into a dynamically created protoregistry.Types.
		extTypes := new(protoregistry.Types)
		for _, file := range gen.Files {
			if err := registerAllExtensions(extTypes, file.Desc); err != nil {
				panic(err)
			}
		}

		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			generateFile(gen, f, extTypes)
		}
		return nil
	})
}

// generateFile generates a _ascii.pb.go file containing gRPC service definitions.
func generateFile(gen *protogen.Plugin, file *protogen.File, extTypes *protoregistry.Types) {
	filename := file.GeneratedFilenamePrefix + "_version.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)

	g.P("// Code generated by protoc-gen-go-version. DO NOT EDIT.")
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()

	// The MessageOptions as provided by protoc does not know about
	// dynamically created extensions, so they are left as unknown fields.
	// We round-trip marshal and unmarshal the options with
	// a dynamically created resolver that does know about extensions at runtime.
	options := file.Desc.Options().(*descriptorpb.FileOptions)
	b, err := proto.Marshal(options)
	if err != nil {
		panic(err)
	}
	options.Reset()
	err = proto.UnmarshalOptions{Resolver: extTypes}.Unmarshal(b, options)
	if err != nil {
		panic(err)
	}

	version := ""

	// Use protobuf reflection to iterate over all the extension fields,
	// looking for the ones that we are interested in.
	options.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if fd.IsExtension() && fd.FullName() == "iyarkov2.chat.api.version" {
			version = v.String()
			return false
		}
		return true
	})

	// Add Version constant
	g.P(fmt.Sprintf("const Version = \"%s.%s\"", version, *rev))
	g.P()

	// Add client-side interceptor
	g.P(fmt.Sprintf("func WithServerVersion() grpc.ServerOption {"))
	g.P(fmt.Sprintf("\treturn version.WithServerInterceptor()"))
	g.P(fmt.Sprintf("}"))

	// Add client-side interceptor
	g.P(fmt.Sprintf("func WithClientVersion() grpc.DialOption {"))
	g.P(fmt.Sprintf("\treturn version.WithClientInterceptor(Version, methods)"))
	g.P(fmt.Sprintf("}"))

	// Add version method to every server side stub
	for _, service := range file.Services {
		serviceName := service.Desc.FullName()
		g.P(fmt.Sprintf("func (Unimplemented%sServer) Version() string {", service.GoName))
		g.P(fmt.Sprintf("\treturn Version"))
		g.P(fmt.Sprintf("}"))

		methods := service.Desc.Methods()

		g.P(fmt.Sprintf(""))
		g.P(fmt.Sprintf("var methods = map[string]bool {"))
		for i := 0; i < methods.Len(); i++ {
			g.P(fmt.Sprintf("\t \"/%s/%s\": true,", serviceName,  methods.Get(i).Name()))
		}
		g.P(fmt.Sprintf("}"))
	}


}

// Recursively register all extensions into the provided protoregistry.Types,
// starting with the protoreflect.FileDescriptor and recursing into its MessageDescriptors,
// their nested MessageDescriptors, and so on.
//
// This leverages the fact that both protoreflect.FileDescriptor and protoreflect.MessageDescriptor
// have identical Messages() and Extensions() functions in order to recurse through a single function
func registerAllExtensions(extTypes *protoregistry.Types, descs interface {
	Messages() protoreflect.MessageDescriptors
	Extensions() protoreflect.ExtensionDescriptors
}) error {
	mds := descs.Messages()
	for i := 0; i < mds.Len(); i++ {
		if err := registerAllExtensions(extTypes, mds.Get(i)) ; err != nil {
			return err
		}
	}
	xds := descs.Extensions()
	for i := 0; i < xds.Len(); i++ {
		if err := extTypes.RegisterExtension(dynamicpb.NewExtensionType(xds.Get(i))); err != nil {
			return err
		}
	}
	return nil
}