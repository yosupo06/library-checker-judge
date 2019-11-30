package main

import (
	"os"
	"strings"
	"google.golang.org/grpc"
	pb "github.com/yosupo06/library-checker-judge/api/proto"
	"testing"
	"context"
	"log"
)

var client pb.LibraryCheckerServiceClient

func TestLangList(t *testing.T) {
	ctx := context.Background()
	list, err := client.LangList(ctx, &pb.LangListRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Langs) == 0 {
		t.Fatal(err)
	}
}

func TestSubmitBig(t *testing.T) {
	ctx := context.Background()
	bigSrc := strings.Repeat("a", 3 * 1000 * 1000) // 3 MB
	_, err := client.Submit(ctx, &pb.SubmitRequest{
		Problem: "aplusb",
		Source: bigSrc,
		Lang: "cpp",
	})	
	if err == nil {
		t.Fatal("Success to submit big source")
	}
	t.Log(err)
}

type loginCreds struct{}

func (c *loginCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	dict := map[string]string{}
	if token, ok := ctx.Value("token").(string); ok && token != "" {
		dict["authorization"] = "bearer " + token
	}
	return dict, nil
}

func (c *loginCreds) RequireTransportSecurity() bool {
	return false
}

func TestMain(m *testing.M) {
	log.Print("UNKO")
	options := []grpc.DialOption{grpc.WithBlock(), grpc.WithPerRPCCredentials(&loginCreds{}), grpc.WithInsecure()}
	conn, err := grpc.Dial("localhost:50051", options...)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	log.Print("UNKO")
	client = pb.NewLibraryCheckerServiceClient(conn)

	os.Exit(m.Run())
}
