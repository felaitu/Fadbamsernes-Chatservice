package main

import (
	proto "Exercise5/grpc"
	"context"
	"log"
	"net"
	"os"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Struct that will be used to represent the Server.
type Server struct {
	proto.UnimplementedMessageServiceServer
	name string
	ip   string
	port int
}

var (
	clients = make([]string, 0)
	t       = 0
)

// Used to get the user-defined port for the server from the command line

func main() {
	// Create a server struct
	server := &Server{
		name: "serverName",
		ip:   os.Getenv("IP"),
		port: 6969,
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
		log.Fatal("Could not serve listener")
	}
}

func logMessageRpc(in *proto.MessageData) {
	for idx, clientIp := range clients {
		if idx == 0 {
			log.Printf("%s\n", in.ClientMessage)
		}

		conn, err := grpc.Dial(clientIp+":"+strconv.Itoa(6969), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to connect to ip %s\n", clientIp)
		}

		client := proto.NewClientServiceClient(conn)

		_, err = client.LogMessage(context.Background(), &proto.MessageData{
			ClientIp:      in.ClientIp,
			ClientMessage: in.ClientMessage,
		})

		if err != nil {
			log.Fatal(err.Error())
		}
	}
}

func registerClient(in *proto.MessageData) {
	clients = append(clients, in.ClientIp)
}

func (c *Server) SendMessageToServer(ctx context.Context, in *proto.MessageData) (*proto.Confirmation, error) {
	logMessageRpc(in)

	return &proto.Confirmation{
		Confirmation: 200,
	}, nil
}

// WIP
func (c *Server) Register(ctx context.Context, in *proto.MessageData) (*proto.Confirmation, error) {
	log.Printf("Received register from client %s\n", in.ClientIp)

	registerClient(in)

	return &proto.Confirmation{
		Confirmation: 200,
	}, nil
}
