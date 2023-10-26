package main

import (
	proto "Exercise5/grpc"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/AlbertRossJoh/itualgs_go/fundamentals/queue"
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
	clients            = make([]string, 0)
	lamport      int64 = 0
	messageQueue       = queue.NewBufQueue[*proto.MessageData](1024)
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

	go messageHandler()

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
			unregisterClient(clientIp)
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
			unregisterClient(clientIp)
			continue
		}
		conn.Close()
	}
}

func messageHandler() {
	for {
		for !messageQueue.IsEmpty() {
			val, _ := messageQueue.Dequeue()
			logMessageRpc(val)
		}
	}
}

func unregisterClient(clientIp string) {
	messageQueue.Enqueue(&proto.MessageData{
		ClientIp:      "Server",
		ClientMessage: fmt.Sprintf("Participant %s left Chitty-Chat at Lamport time %d", clientIp, lamport),
		LamportTs:     lamport,
	})
	tmp := make([]string, 0)
	for _, clientIp1 := range clients {
		if clientIp1 != clientIp {
			tmp = append(tmp, clientIp1)
		}
	}
	clients = tmp
}

func registerClient(in *proto.MessageData) {
	clients = append(clients, in.ClientIp)
	messageQueue.Enqueue(&proto.MessageData{
		ClientIp:      "Server",
		ClientMessage: fmt.Sprintf("Participant %s  joined Chitty-Chat at Lamport time %d\n", in.ClientIp, lamport),
		LamportTs:     lamport,
	})
}

func (c *Server) SendMessageToServer(ctx context.Context, in *proto.MessageData) (*proto.Confirmation, error) {
	if in.LamportTs > lamport {
		lamport = in.LamportTs
	}
	lamport += 1

	messageQueue.Enqueue(in)

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

	registerClient(in)

	return &proto.Confirmation{
		Confirmation: 200,
	}, nil
}
