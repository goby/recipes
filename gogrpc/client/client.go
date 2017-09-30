package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"

	pb "github.com/goby/recipes/gogrpc/chatpb"
)

var (
	tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca_file", "testdata/ca.pem", "The file containning the CA root cert file")
	serverAddr         = flag.String("server_addr", "127.0.0.1:10000", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name use to verify the hostname returned by TLS handshake")
)

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	if *tls {
		var sn string
		if *serverHostOverride != "" {
			sn = *serverHostOverride
		}
		var creds credentials.TransportCredentials
		if *caFile != "" {
			var err error
			creds, err = credentials.NewClientTLSFromFile(*caFile, sn)
			if err != nil {
				grpclog.Fatalf("Failed to create TLS credentials %v", err)
			}
		} else {
			creds = credentials.NewClientTLSFromCert(nil, sn)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	var name string

	fmt.Printf("Please insert your name: ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		name = scanner.Text()
	}

	client := pb.NewChatClient(conn)

	user := &pb.UserInfo{
		ID:   0,
		Name: name,
	}
	param := &pb.JoinRequest{RoomID: 0, User: user}
	ret, err := client.Join(context.Background(), param)

	if err != nil {
		grpclog.Fatalf("Join server failed: %v", err)
	}

	stopCh := make(chan struct{})
	msgCh := make(chan *pb.ChatMessage)
	sendCh := make(chan *pb.ChatMessage)

	go func() {
		for {
			select {
			case <-stopCh:
				return
			case msg := <-msgCh:
				if msg.To == nil {
					grpclog.Printf("@%s: %s", msg.From.Name, msg.Message)
				} else {
					grpclog.Printf("[SEC]@%s: %s", msg.From.Name, msg.Message)
				}
			case msg := <-sendCh:
				client.SendChatMessage(context.Background(), msg)
			}
		}
	}()

	go func() {
		for {
			msg, err := ret.Recv()
			if err == io.EOF {
				stopCh <- struct{}{}
			}
			if err != nil {
				grpclog.Fatalf("Receive failed: %v", err)
			}

			msgCh <- msg
		}
	}()

	for scanner.Scan() {
		text := scanner.Text()

		if text == "q" || text == "quit" {
			stopCh <- struct{}{}
			break
		} else if len(text) > 0 {
			msg := &pb.ChatMessage{Message: text, From: user}
			sendCh <- msg
		}
	}

	fmt.Println("End client ...")
}
