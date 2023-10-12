package main

import (
	proto "Exercise5/grpc"
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/AlbertRossJoh/itualgs_go/fundamentals/queue"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	proto.UnimplementedClientServiceServer
	id int
}

var (
	port         = flag.Int("cPort", 0, "client port number")
	serverPort   = flag.Int("sPort", 0, "server port number")
	messageQueue = queue.NewQueue[proto.MessageData](1024)
)

func main() {
	// Parse the flags to get the port for the client
	flag.Parse()

	// Create a client
	client := &Client{
		id: *port,
	}

	// Wait for the client (user) to ask for the time
	go sendMessageFromStdin(client)
	go getMessagesFromServer()

	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(client.id))
	if err != nil {
		log.Fatalf("Could not create the server %v", err)
	}
	proto.RegisterClientServiceServer(grpcServer, client)
	grpcServer.Serve(listener)

	for {

	}
}

func connectToServer() (proto.MessageServiceClient, error) {
	// Dial the server at the specified port.
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(*serverPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to port %d", *serverPort)
	} else {
		log.Printf("Connected to the server at port %d\n", *serverPort)
	}
	return proto.NewMessageServiceClient(conn), nil
}

func sendMessageFromStdin(client *Client) {
	// Connect to the server
	serverConnection, _ := connectToServer()

	// Wait for input in the client terminal
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		log.Printf("Client asked for time with input: %s\n", input)

		// Ask the server for the time
		confirmation, err := serverConnection.SendMessageToServer(context.Background(), &proto.MessageData{
			FromClientId:  int64(client.id),
			Recipient:     int64(client.id), // crazy
			ClientMessage: input,
		})

		if err != nil {
			log.Printf(err.Error())
		} else {
			log.Printf("Confirmation received from the server: %d\n", confirmation.Confirmation)
		}
	}
}

func getMessagesFromServer() error {
	for {
		for !messageQueue.IsEmpty() {
			currentMessage, err := messageQueue.Dequeue()
			if err != nil {
				log.Printf(err.Error())
			} else {
				fmt.Printf("\n %d: %s\n", currentMessage.FromClientId, currentMessage.ClientMessage)
			}
		}
	}
}

func (c *Client) LogMessage(ctx context.Context, in *proto.MessageData) (*proto.Confirmation, error) {
	log.Printf("Received message from server : %s\n", in.ClientMessage)
	messageQueue.Enqueue(*in)
	return &proto.Confirmation{
		Confirmation: 200,
	}, nil
}
