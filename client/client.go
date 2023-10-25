package main

import (
	proto "Exercise5/grpc"
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/AlbertRossJoh/itualgs_go/fundamentals/queue"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var NpcDialogOptions = []string{
	"Have you heard of the high elves?",
	"Hey you. You're finally awake.",
	"Thank you kind sir.",
	"Stop at once. You're breaking the law.",
	"Then serve with your blood.",
	"Have you heard the tragedy of Darth Plaguis the wise?",
	"Jones BBQ and Foot massage! Is my absolute favorite place to go out",
	"Have you ever seen the rain?",
	"According to all known laws of aviation, there is no way a bee should be able to fly",
	"After all this time?",
	"Somebody once told me world was gonna roll me.",
	"I ain't the sharpest tool in the shed.",
	"It's over 9000!!!!",
	"Womp womp",
	"Wouldn't you like to know, weatherboy",
	"It's Hurricane Katrina, more like Hurricane Tortilla",
	"To be or not to be, that is the question.",
	"It was the best of times, it was the worst of times.",
	"All happy families are alike; each unhappy family is unhappy in its own way.",
	"The only way to deal with all this meaningless of life was to find a distraction.",
	"There is nothing either good or bad, but thinking makes it so.",
	"Call me Ishmael.",
	"Happy families are all alike; every unhappy family is unhappy in its own way.",
	"It was a bright cold day in April, and the clocks were striking thirteen.",
	"I have nothing to declare except my genius.",
	"You can't wait for inspiration. You have to go after it with a club.",
	"It is not a lack of love, but a lack of friendship that makes unhappy marriages.",
	"And, when you want something, all the universe conspires in helping you to achieve it.",
}

type Client struct {
	proto.UnimplementedClientServiceServer
	ip string
}

const SERVER_PORT = 6969

var (
	serverIp           = flag.String("serverIpAddr", "172.20.0.100", "server ip")
	messageQueue       = queue.NewQueue[proto.MessageData](1024)
	lamport      int64 = 0
)

func main() {
	// Parse the flags to get the port for the client
	flag.Parse()

	// Create a client
	client := &Client{
		ip: os.Getenv("HOSTNAME"),
	}

	// Wait for the client (user) to ask for the time
	//go sendMessageFromStdin(client)
	go sendMessagesWithInterval(client)
	go getMessagesFromServer()

	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(SERVER_PORT))
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
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", *serverIp, strconv.Itoa(SERVER_PORT)), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to port %d", SERVER_PORT)
	}

	return proto.NewMessageServiceClient(conn), nil
}

func registerWithServer(serverConnection proto.MessageServiceClient, client *Client) (*proto.Confirmation, error) {
	lamport++
	return serverConnection.Register(context.Background(), &proto.MessageData{
		ClientIp:      client.ip,
		ClientMessage: "Register",
		LamportTs:     lamport,
	})
}

func sendMessageFromStdin(client *Client) {
	// Connect to the server
	serverConnection, _ := connectToServer()

	_, err := registerWithServer(serverConnection, client)

	if err != nil {
		log.Fatal("Failed to connect to the server!\n", err)
	}

	// Wait for input in the client terminal
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()

		lamport++
		// Ask the server for the time
		_, err := serverConnection.SendMessageToServer(context.Background(), &proto.MessageData{
			ClientIp:      client.ip,
			ClientMessage: input,
			LamportTs:     lamport,
		})

		if err != nil {
			log.Fatal(err.Error())
		}
	}
}

// This will prompt the docker compose to re-join
func leaveServer() {
	os.Exit(0)
}

func sendMessagesWithInterval(client *Client) {
	// Connect to the server
	serverConnection, _ := connectToServer()

	_, err := registerWithServer(serverConnection, client)

	if err != nil {
		log.Fatal("Failed to connect to the server!\n", err)
	}

	// Wait for input in the client terminal
	for {
		time.Sleep(time.Duration(rand.Intn(15)) * time.Second)

		if rand.Intn(len(NpcDialogOptions)) == 2 {
			leaveServer()
		}

		randomDialog := NpcDialogOptions[rand.Intn(len(NpcDialogOptions))]
		textMessage := fmt.Sprintf("[%s]: %s", client.ip, randomDialog)

		lamport++
		// Ask the server for the time
		_, err := serverConnection.SendMessageToServer(context.Background(), &proto.MessageData{
			ClientIp:      client.ip,
			ClientMessage: textMessage,
			LamportTs:     lamport,
		})

		if err != nil {
			log.Fatal(err.Error())
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
				fmt.Printf("\n\n\n%s %d\n\n\n", currentMessage.ClientMessage, currentMessage.LamportTs)
			}
		}
	}
}

func (c *Client) LogMessage(ctx context.Context, in *proto.MessageData) (*proto.Confirmation, error) {
	if in.LamportTs > lamport {
		lamport = in.LamportTs
	}
	lamport += 1

	messageQueue.Enqueue(*in)
	return &proto.Confirmation{
		Confirmation: 200,
	}, nil
}
