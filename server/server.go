package main

import (
	proto "Exercise5/grpc"
	"context"
	"flag"
	"log"
	"net"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Struct that will be used to represent the Server.
type Server struct {
	proto.UnimplementedMessageServiceServer
	name string
	port int
}

// Used to get the user-defined port for the server from the command line
var port = flag.Int("port", 0, "server port number")

func main() {
	// Get the port from the command line when the server is run
	flag.Parse()

	// Create a server struct
	server := &Server{
		name: "serverName",
		port: *port,
	}

	// Start the server
	go startServer(server)

	// Keep the server running until it is manually quit
	for {

	}
}

func startServer(server *Server) {

	// Create a new grpc server
	grpcServer := grpc.NewServer()

	// Make the server listen at the given port (convert int port to string)
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(server.port))

	if err != nil {
		log.Fatalf("Could not create the server %v", err)
	}
	log.Printf("Started server at port: %d\n", server.port)

	// Register the grpc server and serve its listener
	proto.RegisterMessageServiceServer(grpcServer, server)
	serveError := grpcServer.Serve(listener)
	if serveError != nil {
		log.Fatalf("Could not serve listener")
	}
}

func logMessageRpc(in *proto.MessageData) {
	clientPort := int(in.Recipient)

	conn, err := grpc.Dial("localhost:"+strconv.Itoa(clientPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Womp womp")
	} else {
		log.Printf("Connected to the client at port %d\n", clientPort)
	}

	client := proto.NewClientServiceClient(conn)

	log.Printf("[+] Sending message %s from %d to %d", in.ClientMessage, in.FromClientId, in.Recipient)
	client.LogMessage(context.Background(), &proto.MessageData{
		FromClientId:  int64(clientPort),
		Recipient:     int64(clientPort), // crazy
		ClientMessage: in.ClientMessage,
	})
}

func (c *Server) SendMessageToServer(ctx context.Context, in *proto.MessageData) (*proto.Confirmation, error) {
	log.Printf("Received message from client %d : %s\n", in.FromClientId, in.ClientMessage)

	logMessageRpc(in)

	return &proto.Confirmation{
		Confirmation: 200,
	}, nil
}
