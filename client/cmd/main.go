package main

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
)

func registerProtoFile(srcDir string, filename string) error {
	// First, convert the .proto file to a file descriptor set.
	tmpFile := path.Join(srcDir, filename + ".pb")
	cmd := exec.Command("protoc",
		"-I " + srcDir,
		"--include_source_info",
		"--descriptor_set_out=" + tmpFile,
		"--proto_path="+srcDir,
		path.Join(srcDir, filename))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "protoc")
	}

	defer os.Remove(tmpFile)

	// Now load that temporary file as a file descriptor set protobuf.
	protoFile, err := ioutil.ReadFile(tmpFile)
	if err != nil {
		return errors.Wrapf(err, "read tmp file")
	}

	pbSet := new(descriptorpb.FileDescriptorSet)
	if err := proto.Unmarshal(protoFile, pbSet); err != nil {
		return errors.Wrapf(err, "unmarshal")
	}

	// We know protoc was invoked with a single .proto file.
	pb := pbSet.GetFile()[0]

	// Initialize the file descriptor object.
	fd, err := protodesc.NewFile(pb, protoregistry.GlobalFiles)
	if err != nil {
		return errors.Wrapf(err, "NewFile")
	}

	// and finally register it.
	return protoregistry.GlobalFiles.RegisterFile(fd)
}

func main() {

	/*
		Step 1 - Register all known proto files with protoregistry.
	    NOTES
			- this is a copy of an example from the internet. It uses protoc to compile proto file into pb descriptor file.
			  Another way to do it is to compile proto files into pb files BEFORE we feed them to the app
			- Looks like google.golang.org/grpc registers all google/protobuf's files during initialization. I leave them
			  commented for the reference
	 */
	{
		fmt.Println("\n----------------------- Registering proto files ------------------------------------\n")
		files := []string{
			//"google/protobuf/any.proto",
			//"google/protobuf/duration.proto",
			//"google/protobuf/empty.proto",
			//"google/protobuf/field_mask.proto",
			//"google/protobuf/timestamp.proto",
			//"google/rpc/status.proto",
			//"google/type/expr.proto",
			//"google/api/http.proto",
			//"google/api/annotations.proto",
			//"google/api/client.proto",
			//"google/api/field_behavior.proto",
			//"google/api/resource.proto",
			//"google/iam/v1/options.proto",
			//"google/iam/v1/policy.proto",
			//"google/iam/v1/iam_policy.proto",
			//"google/longrunning/operations.proto",
			//"google/cloud/functions/v1/functions.proto",
			"chat.proto",
		}
		for _, f := range files {
			err := registerProtoFile("../api", f)
			if err != nil {
				panic(errors.Wrapf(err, f))
			}
		}
	}

	/*
		Uncomment lines below if you would like to peek inside protoregistry
	*/
	//protoregistry.GlobalFiles.RangeFiles(func (fd protoreflect.FileDescriptor) bool {
	//	fmt.Printf("%v\n\n", fd)
	//	return true
	//})

	/*
		Step 2.a - read binary message from a file. It could be any binary blob or stream - a database, Kafka message, etc
		We need to know the name of the message
		Fields are accessible by names, or we can iterate over all names and convert a proto object into a flat map
	 */
	{
		fmt.Println("\n----------------------- Reading from out/message.out ------------------------------------\n")
		d, err := protoregistry.GlobalFiles.FindDescriptorByName("iyarkov2.chat.api.ConnectRequest")
		if err != nil {
			panic(errors.Wrapf(err, "Descriptor not found"))
		}
		fmt.Printf("\u001B[32mDescriptor\u001B[0m %v\n", d)

		fmt.Printf("Reading from a file\n")
		in, err := ioutil.ReadFile("../out/message.out")
		if err != nil {
			log.Fatalln("Error reading file:", err)
		}
		message := dynamicpb.NewMessage(d.(protoreflect.MessageDescriptor))
		if err := proto.Unmarshal(in, message); err != nil {
			log.Fatalln("Failed to parse message file:", err)
		}
		fmt.Printf("Entire message: [%v]\n", message)

		nameField := message.Descriptor().Fields().ByName("name")
		value := message.Get(nameField)
		fmt.Printf("Sinle field: [%s]\n", value)
	}

	/*
		Step 2.b - making gRPC call. To make that call the client need to know 4 parameters:
		* The server name
		* The request message name
		* The response message name
		* The server method name
	*/

	{
		fmt.Println("\n----------------------- gRPC call -------------------------------------------------------\n")
		serviceDesc, err := protoregistry.GlobalFiles.FindDescriptorByName("iyarkov2.chat.api.ChatService")
		if err != nil {
			log.Fatalln("Service not found:", err)
		}
		fmt.Printf("\u001b[32mService Descriptor\u001B[0m  %v\n\n", serviceDesc)

		//var opts []grpc.DialOption
		conn, err := grpc.Dial("localhost:8888", grpc.WithInsecure())
		if err != nil {
			log.Fatalln("Connection error:", err)
		}
		defer conn.Close()

		requestD, err := protoregistry.GlobalFiles.FindDescriptorByName("iyarkov2.chat.api.ConnectRequest")
		if err != nil {
			panic(errors.Wrapf(err, "Request Descriptor not found"))
		}
		fmt.Printf("\u001B[32mIn Message Descriptor\u001B[0m %v\n\n", requestD)
		request := dynamicpb.NewMessage(requestD.(protoreflect.MessageDescriptor))
		requestNameField := request.Descriptor().Fields().ByName("name")
		request.Set(requestNameField, protoreflect.ValueOfString("John"))

		responseD, err := protoregistry.GlobalFiles.FindDescriptorByName("iyarkov2.chat.api.ConnectResponse")
		if err != nil {
			panic(errors.Wrapf(err, "Response Descriptor not found"))
		}
		fmt.Printf("\u001B[32mOut Message Descriptor\u001B[0m %v\n\n", responseD)
		response := dynamicpb.NewMessage(responseD.(protoreflect.MessageDescriptor))

		invokeError := conn.Invoke(context.Background(), "/iyarkov2.chat.api.ChatService/Connect", request, response)
		if invokeError != nil {
			panic(errors.Wrapf(invokeError, "Invoke error"))
		}
		fmt.Printf("Response: [%v]\n", response)

		idField := response.Descriptor().Fields().ByName("user_id")
		value := response.Get(idField)
		fmt.Printf("Sinle field: [%s]\n", value)

		// What if a field does not exist?
		fieldDoesNotExist := response.Descriptor().Fields().ByName("does_not_exist")
		if fieldDoesNotExist == nil {
			fmt.Println("field not found - as expected")
		} else {
			panic("field found")
		}

		// DO not uncomment - will cause panic: runtime error
		//response.Get(fieldDoesNotExist)
	}

}
