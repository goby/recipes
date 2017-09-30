package main

import (
	"flag"
	"fmt"
	"net"
	"sync"
	"time"

	context "golang.org/x/net/context"

	pb "github.com/goby/recipes/gogrpc/chatpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

var (
	tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile   = flag.String("cert_file", "testdata/server1.pem", "The TLS cert file")
	keyFile    = flag.String("key_file", "testdata/server1.key", "The TLS key file")
	jsonDBFile = flag.String("json_db_file", "testdata/route_guide_db.json", "A json file containing a list of features")
	port       = flag.Int("port", 10000, "The server port")
)

type chatServer struct {
	users map[string]pb.Chat_JoinServer
	mu    sync.Mutex
}

type clientChannel struct {
	pb.Chat_JoinServer
}

func (s *chatServer) registerClient(id string, c clientChannel) error {
	return nil
}

// A client-to-server rpc
func (s *chatServer) Hello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	return nil, nil

}

// A server-to-client rpc, join a room and register chat client
func (s *chatServer) Join(req *pb.JoinRequest, stream pb.Chat_JoinServer) error {
	user := req.GetUser()
	if _, ok := s.users[user.GetName()]; ok {
		grpclog.Printf("User %v already join, ignored.", user.GetName())
		return fmt.Errorf("multiple join")
	}

	c := &clientChannel{stream}

	s.mu.Lock()
	s.users[user.GetName()] = c
	s.mu.Unlock()

	grpclog.Printf("User %v join to server.", user.GetName())

	<-c.Context().Done()
	grpclog.Printf("user %v left from server", user.GetName())

	if err := c.Context().Err(); err != nil && err != context.Canceled {
		grpclog.Printf("Client failed: %v.", err)
	}

	s.mu.Lock()
	delete(s.users, user.GetName())
	s.mu.Unlock()

	return nil
}

// A client-to-server rpc, send message to server
func (s *chatServer) SendChatMessage(ctx context.Context, msg *pb.ChatMessage) (*pb.SendResponse, error) {
	grpclog.Printf("Recv from client: %s", msg.Message)

	if msg.From == nil {
		return nil, fmt.Errorf("Unkown from")
	}

	if msg.To == nil {
		s.mu.Lock()
		for name, user := range s.users {
			if name != msg.From.Name {
				if err := user.Send(msg); err != nil {
					grpclog.Printf("send failed: %v", err)
				}
			}
		}
		s.mu.Unlock()
	}

	return &pb.SendResponse{Code: 0, Message: "OK"}, nil
}

// Not implementation
func (s *chatServer) Chat(stream pb.Chat_ChatServer) error {
	return nil
}

func (s *chatServer) sendMessage() error {
	for name, user := range s.users {
		msg := &pb.ChatMessage{Message: name}

		if err := user.Send(msg); err != nil {
			grpclog.Printf("send failed: %v", err)
		}
	}
	return nil
}

func (s *chatServer) generateMessage(stopCh <-chan struct{}) error {

	t := time.Tick(1 * time.Second)

	for {
		select {
		case <-t:
			s.sendMessage()
		case <-stopCh:
			break
		}
	}

	return nil

}

func newServer() *chatServer {
	s := new(chatServer)
	s.users = make(map[string]pb.Chat_JoinServer)

	//stopCh := make(chan struct{})

	//go s.generateMessage(stopCh)

	return s
}

// The entry point of the server
func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		grpclog.Fatalf("Listen failed: %v", err)
	}

	grpclog.Printf("Listen on port %d", *port)

	var opts []grpc.ServerOption
	if *tls {
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			grpclog.Fatalf("Create credentials failed: %v", err)
		}

		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterChatServer(grpcServer, newServer())
	if err := grpcServer.Serve(lis); err != nil {
		grpclog.Fatalf("Serve on %v failed: %v", lis, err)
	}
}
