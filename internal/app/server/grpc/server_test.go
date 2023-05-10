package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	"github.com/blokhinnv/shorty/internal/app/log"
	"github.com/blokhinnv/shorty/internal/app/server/config"
	"github.com/blokhinnv/shorty/internal/app/storage"
	pb "github.com/blokhinnv/shorty/proto"
	"github.com/golang/mock/gomock"
)

type GRPCTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller
	db   *storage.MockStorage
}

func (suite *GRPCTestSuite) SetupSuite() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.db = storage.NewMockStorage(suite.ctrl)
}

func (suite *GRPCTestSuite) server(ctx context.Context) (pb.ShortyClient, func()) {
	// src: https://medium.com/@3n0ugh/how-to-test-grpc-servers-in-go-ba90fe365a18
	buffer := 101024 * 1024
	lis := bufconn.Listen(buffer)

	srvCloseCh := make(chan struct{}, 1)
	srvImpl := NewShortyServer(
		suite.db,
		"http://localhost:8080",
		[]byte("shorty"),
		"192.168.0.0/24",
		time.Millisecond,
		srvCloseCh,
	)

	baseServer := grpc.NewServer(
		withServerUnaryInterceptor(srvImpl),
		withServerStreamInterceptor(srvImpl),
	)

	pb.RegisterShortyServer(baseServer, srvImpl)
	go func() {
		if err := baseServer.Serve(lis); err != nil {
			log.Printf("error serving server: %v", err)
		}
	}()

	conn, err := grpc.DialContext(ctx, "",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("error connecting to server: %v", err)
	}

	closer := func() {
		err := lis.Close()
		if err != nil {
			log.Printf("error closing listener: %v", err)
		}
		baseServer.Stop()
	}

	client := pb.NewShortyClient(conn)

	return client, closer
}

func (suite *GRPCTestSuite) TestGetOriginalURL() {
	ctx := context.Background()
	client, closer := suite.server(ctx)
	defer closer()
	suite.T().Run("OK", func(t *testing.T) {
		suite.db.EXPECT().
			GetURLByID(gomock.Any(), "qwerty").
			Return(storage.Record{URL: "shorty.com"}, nil)
		in := &pb.GetOriginalURLRequest{UrlId: "qwerty"}
		out, err := client.GetOriginalURL(ctx, in)
		suite.NoError(err)
		suite.Equal("shorty.com", out.Url)
	})

	suite.T().Run("Deleted", func(t *testing.T) {
		suite.db.EXPECT().
			GetURLByID(gomock.Any(), "qwerty").
			Return(storage.Record{}, storage.ErrURLWasDeleted)
		in := &pb.GetOriginalURLRequest{UrlId: "qwerty"}
		_, err := client.GetOriginalURL(ctx, in)
		suite.Error(err)
	})

	suite.T().Run("BadURL", func(t *testing.T) {
		in := &pb.GetOriginalURLRequest{UrlId: "@!%"}
		_, err := client.GetOriginalURL(ctx, in)
		suite.Error(err)
	})
}

func (suite *GRPCTestSuite) TestGetShortURL() {
	ctx := context.Background()
	client, closer := suite.server(ctx)
	defer closer()
	suite.T().Run("OK", func(t *testing.T) {
		suite.db.EXPECT().
			AddURL(gomock.Any(), "http://qwerty.com", gomock.Any(), gomock.Any()).
			Return(nil)
		in := &pb.GetShortURLRequest{Url: "http://qwerty.com"}
		out, err := client.GetShortURL(ctx, in)
		suite.NoError(err)
		suite.Equal("http://localhost:8080/nrt9sg4_2feb2", out.Url)
	})

	suite.T().Run("Error", func(t *testing.T) {
		suite.db.EXPECT().
			AddURL(gomock.Any(), "http://qwerty.com", gomock.Any(), gomock.Any()).
			Return(storage.ErrUniqueViolation)
		in := &pb.GetShortURLRequest{Url: "http://qwerty.com"}
		_, err := client.GetShortURL(ctx, in)
		suite.Error(err)
	})
}

func (suite *GRPCTestSuite) TestGetOriginalURLs() {
	ctx := context.Background()
	client, closer := suite.server(ctx)
	defer closer()
	suite.T().Run("Error", func(t *testing.T) {
		suite.db.EXPECT().
			GetURLsByUser(gomock.Any(), gomock.Any()).
			Return(nil, fmt.Errorf("some error..."))
		in := &pb.GetOriginalURLsRequest{}
		out, _ := client.GetOriginalURLs(ctx, in)
		_, err := out.Recv()
		suite.Error(err)
	})

	suite.T().Run("OK", func(t *testing.T) {
		expected := []storage.Record{{URL: "testURL", URLID: "testURLID"}}
		suite.db.EXPECT().
			GetURLsByUser(gomock.Any(), gomock.Any()).
			Return(expected, nil)
		in := &pb.GetOriginalURLsRequest{}
		out, err := client.GetOriginalURLs(ctx, in)
		suite.NoError(err)

		var outs []*pb.GetOriginalURLsResponse
		for {
			o, err := out.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			outs = append(outs, o)
		}
		for i, o := range outs {
			suite.Equal(expected[i].URL, o.Url)
			suite.Equal(expected[i].URLID, o.UrlId)
		}
	})
}

func (suite *GRPCTestSuite) TestGetShortURLJSON() {
	ctx := context.Background()
	client, closer := suite.server(ctx)
	defer closer()
	suite.T().Run("OK", func(t *testing.T) {
		suite.db.EXPECT().
			AddURL(gomock.Any(), "http://qwerty.com", gomock.Any(), gomock.Any()).
			Return(nil)
		in := &pb.GetShortURLJSONRequest{
			Item: &pb.GetShortURLJSONRequest_Item{Url: "http://qwerty.com"},
		}
		out, err := client.GetShortURLJSON(ctx, in)
		suite.NoError(err)
		suite.Equal("http://localhost:8080/nrt9sg4_2feb2", out.Item.Result)
	})

	suite.T().Run("Error", func(t *testing.T) {
		suite.db.EXPECT().
			AddURL(gomock.Any(), "http://qwerty.com", gomock.Any(), gomock.Any()).
			Return(storage.ErrUniqueViolation)
		in := &pb.GetShortURLJSONRequest{
			Item: &pb.GetShortURLJSONRequest_Item{Url: "http://qwerty.com"},
		}
		_, err := client.GetShortURLJSON(ctx, in)
		suite.Error(err)
	})
}

func (suite *GRPCTestSuite) TestGetShortURLBatch() {
	ctx := context.Background()
	client, closer := suite.server(ctx)
	defer closer()
	suite.T().Run("OK", func(t *testing.T) {
		suite.db.EXPECT().
			AddURLBatch(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)

		in := &pb.GetShortURLBatchRequest{Batch: []*pb.GetShortURLBatchRequest_Item{
			{CorrelationId: "test1", OriginalUrl: "https://mail.ru/"},
		}}
		out, err := client.GetShortURLBatch(ctx, in)
		suite.NoError(err)
		suite.Equal("test1", out.Batch[0].CorrelationId)
		suite.Equal("http://localhost:8080/f3o7hcrcrupz1", out.Batch[0].ShortUrl)
	})

	suite.T().Run("Error", func(t *testing.T) {
		suite.db.EXPECT().
			AddURLBatch(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(storage.ErrUniqueViolation)
		in := &pb.GetShortURLBatchRequest{Batch: []*pb.GetShortURLBatchRequest_Item{
			{CorrelationId: "test1", OriginalUrl: "https://mail.ru/"},
		}}
		_, err := client.GetShortURLBatch(ctx, in)
		suite.Error(err)
	})
}

func (suite *GRPCTestSuite) TestGetStats() {
	ctx := context.Background()
	client, closer := suite.server(ctx)
	defer closer()
	suite.T().Run("OK", func(t *testing.T) {
		suite.db.EXPECT().
			GetStats(gomock.Any()).
			Return(1, 1, nil)
		md := metadata.New(map[string]string{"X-Real-IP": "192.168.0.1"})
		mdCtx := metadata.NewOutgoingContext(context.Background(), md)
		in := new(pb.GetStatsRequest)
		out, err := client.GetStats(mdCtx, in)
		suite.NoError(err)
		suite.Equal(1, int(out.Users))
		suite.Equal(1, int(out.Urls))
	})

	suite.T().Run("Not contains", func(t *testing.T) {
		md := metadata.New(map[string]string{"X-Real-IP": "192.168.10.11"})
		mdCtx := metadata.NewOutgoingContext(context.Background(), md)
		in := new(pb.GetStatsRequest)
		_, err := client.GetStats(mdCtx, in)
		suite.Error(err)
	})

	suite.T().Run("Bad IP", func(t *testing.T) {
		md := metadata.New(map[string]string{"X-Real-IP": "192.168"})
		mdCtx := metadata.NewOutgoingContext(context.Background(), md)
		in := new(pb.GetStatsRequest)
		_, err := client.GetStats(mdCtx, in)
		suite.Error(err)
	})
}

func (suite *GRPCTestSuite) TestPing() {
	ctx := context.Background()
	client, closer := suite.server(ctx)
	defer closer()
	suite.T().Run("OK", func(t *testing.T) {
		suite.db.EXPECT().
			Ping(gomock.Any()).
			Return(true)
		in := new(pb.PingRequest)
		out, err := client.Ping(ctx, in)
		suite.NoError(err)
		suite.Equal(true, out.Pinged)
	})
}

func (suite *GRPCTestSuite) TestDeleteURL() {
	ctx := context.Background()
	client, closer := suite.server(ctx)
	defer closer()
	suite.T().Run("OK", func(t *testing.T) {
		suite.db.EXPECT().
			DeleteMany(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)
		in := []*pb.DeleteURLRequest{
			{Url: "f3o7hcrcrupz1"},
		}
		outClient, err := client.DeleteURL(ctx)

		for _, v := range in {
			err := outClient.Send(v)
			suite.NoError(err)
		}
		suite.NoError(err)
		_, err = outClient.CloseAndRecv()
		suite.NoError(err)
		time.Sleep(500 * time.Millisecond)
	})
}

func (suite *GRPCTestSuite) TestRunGRPCServer() {
	cfg := &config.ServerConfig{
		SQLiteDBPath:       "demo.sqlte3",
		SQLiteClearOnStart: false,
	}
	ctx, cancel := context.WithCancel(context.Background())
	go RunGRPCServer(ctx, cfg)
	cancel()
}

func TestGRPCTestSuite(t *testing.T) {
	suite.Run(t, new(GRPCTestSuite))
}
