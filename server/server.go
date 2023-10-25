package main

import (
	proto "Exercise5/grpc"
	"context"
	"fmt"
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

const SERVER_PORT = 6969

var (
	clients       = make([]string, 0)
	lamport int64 = 0
)

// Used to get the user-defined port for the server from the command line

func main() {
	// Create a server struct
	server := &Server{
		name: "serverName",
		ip:   os.Getenv("IP"),
		port: SERVER_PORT,
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
			log.Printf("(LT: %d)\t %s\n", in.LamportTs, in.ClientMessage)
		}

		conn, err := grpc.Dial(fmt.Sprintf("%s:%s", clientIp, strconv.Itoa(SERVER_PORT)), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			tmp := make([]string, 0)
			for _, clientIp1 := range clients {
				if clientIp1 != clientIp {
					tmp = append(tmp, clientIp1)
				}
			}
			clients = tmp
			log.Printf("Failed to connect to client with IP: %s\n", clientIp)
			continue
		}

		client := proto.NewClientServiceClient(conn)

		lamport++
		_, err = client.LogMessage(context.Background(), &proto.MessageData{
			ClientIp:      in.ClientIp,
			ClientMessage: in.ClientMessage,
			LamportTs:     lamport,
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

	if in.LamportTs > lamport {
		lamport = in.LamportTs
	}
	lamport += 1

	logMessageRpc(in)

	return &proto.Confirmation{
		Confirmation: 200,
	}, nil
}

// WIP
func (c *Server) Register(ctx context.Context, in *proto.MessageData) (*proto.Confirmation, error) {

	if in.LamportTs > lamport {
		lamport = in.LamportTs
	}
	lamport += 1

	log.Printf("Participant %s  joined Chitty-Chat at Lamport time %d\n", in.ClientIp, lamport)

	registerClient(in)

	return &proto.Confirmation{
		Confirmation: 200,
	}, nil
}
