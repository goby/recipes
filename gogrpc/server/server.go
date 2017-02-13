package main

import (
	"flag"
	"fmt"
	"net"

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
	s.users[user.GetName()] = c
	go func() {
		<-c.Context().Done()
		grpclog.Printf("Client disconnect to server: %v", c.Context())

		if err := c.Context().Err(); err != nil {
			grpclog.Printf("Client failed: %v", err)
		}
		delete(s.users, user.GetName())
	}()

	grpclog.Printf("User %v join to server.", user.GetName())

	return nil
}

// A client-to-server rpc, send message to server
func (s *chatServer) SendChatMessage(ctx context.Context, stream *pb.ChatMessage) (*pb.SendResponse, error) {
	return nil, nil
}

// Not implementation
func (s *chatServer) Chat(stream pb.Chat_ChatServer) error {
	return nil
}

func newServer() *chatServer {
	s := new(chatServer)
	s.users = make(map[string]pb.Chat_JoinServer)
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
