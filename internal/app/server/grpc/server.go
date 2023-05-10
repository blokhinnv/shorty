// Package grpc contains all the logic for gRPC server.
package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/blokhinnv/shorty/internal/app/log"
	"github.com/blokhinnv/shorty/internal/app/shorten"
	"github.com/blokhinnv/shorty/internal/app/storage"
	pb "github.com/blokhinnv/shorty/proto"
)

// ShortyServer supports all necessary server methods.
type ShortyServer struct {
	// one need to embed the type pb.Unimplemented<TypeName>
	// for compatibility with future versions
	pb.UnimplementedShortyServer
	s              storage.Storage
	baseURL        string
	secretKey      []byte
	trustedSubnet  *net.IPNet
	deleteJobs     map[uint32][]string
	m              sync.Mutex
	expireDuration time.Duration
	srvCloseCh     chan struct{}
}

// NewShortyServer is a constructor for ShortyServer.
func NewShortyServer(
	s storage.Storage,
	baseURL string,
	secretKey []byte,
	trustedSubnet string,
	expireDuration time.Duration,
	srvCloseCh chan struct{},
) *ShortyServer {
	srvImpl := ShortyServer{
		s:              s,
		baseURL:        baseURL,
		secretKey:      secretKey,
		expireDuration: expireDuration,
		srvCloseCh:     srvCloseCh,
	}
	if trustedSubnet != "" {
		_, ipv4Net, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			log.Fatalln(err)
		}
		srvImpl.trustedSubnet = ipv4Net
	}
	srvImpl.deleteJobs = make(map[uint32][]string)
	go srvImpl.deleteURLLoop()
	return &srvImpl
}

// GetOriginalURL is a method to retrieve original URL.
func (srv *ShortyServer) GetOriginalURL(
	ctx context.Context,
	req *pb.GetOriginalURLRequest,
) (*pb.GetOriginalURLResponse, error) {
	var response pb.GetOriginalURLResponse

	re := regexp.MustCompile(`^\w+$`)
	if !re.MatchString(req.UrlId) {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect URL")
	}

	rec, err := srv.s.GetURLByID(ctx, req.UrlId)
	if err != nil {
		if errors.Is(err, storage.ErrURLWasDeleted) {
			return nil, status.Errorf(codes.Unavailable, err.Error())
		}
		return nil, status.Errorf(codes.NotFound, err.Error())
	}
	response.Url = rec.URL
	return &response, nil
}

// GetShortURL is a method to retrieve short URL.
func (srv *ShortyServer) GetShortURL(
	ctx context.Context,
	req *pb.GetShortURLRequest,
) (*pb.GetShortURLResponse, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}
	var response pb.GetShortURLResponse
	shortURLID, shortenURL, err := shorten.GetShortURL(req.Url, uint32(userID), srv.baseURL)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	err = srv.s.AddURL(ctx, req.Url, shortURLID, uint32(userID))
	if err != nil {
		if errors.Is(err, storage.ErrUniqueViolation) {
			response.Url = shortenURL
			return &response, status.Errorf(codes.AlreadyExists, err.Error())
		} else {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
	}
	response.Url = shortenURL
	return &response, nil
}

// GetOriginalURLs is a method to all URL that was shortened by user.
func (srv *ShortyServer) GetOriginalURLs(
	req *pb.GetOriginalURLsRequest,
	out pb.Shorty_GetOriginalURLsServer,
) error {
	ctx := out.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		return err
	}
	records, err := srv.s.GetURLsByUser(ctx, userID)
	if err != nil {
		return status.Errorf(codes.NotFound, err.Error())
	}
	for _, rec := range records {
		resp := pb.GetOriginalURLsResponse{Url: rec.URL, UrlId: rec.URLID}
		if err := out.Send(&resp); err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

// GetShortURLJSON is a method to retrieve short URL with extra nested struct.
func (srv *ShortyServer) GetShortURLJSON(
	ctx context.Context,
	req *pb.GetShortURLJSONRequest,
) (*pb.GetShortURLJSONResponse, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}
	var response pb.GetShortURLJSONResponse
	shortURLID, shortenURL, err := shorten.GetShortURL(req.Item.Url, uint32(userID), srv.baseURL)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	err = srv.s.AddURL(ctx, req.Item.Url, shortURLID, uint32(userID))
	if err != nil {
		if errors.Is(err, storage.ErrUniqueViolation) {
			response.Item = &pb.GetShortURLJSONResponse_Item{Result: shortenURL}
			return &response, status.Errorf(codes.AlreadyExists, err.Error())
		} else {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
	}
	response.Item = &pb.GetShortURLJSONResponse_Item{Result: shortenURL}
	return &response, nil
}

// GetShortURLBatch is a method to retrieve short URL for a batch.
func (srv *ShortyServer) GetShortURLBatch(
	ctx context.Context,
	req *pb.GetShortURLBatchRequest,
) (*pb.GetShortURLBatchResponse, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	urlIDs := make(map[string]string)
	result := make([]*pb.GetShortURLBatchResponse_Item, 0, len(req.Batch))
	for _, item := range req.Batch {
		shortURLID, shortenURL, err := shorten.GetShortURL(
			item.OriginalUrl,
			userID,
			srv.baseURL,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		urlIDs[item.OriginalUrl] = shortURLID
		result = append(
			result,
			&pb.GetShortURLBatchResponse_Item{
				CorrelationId: item.CorrelationId,
				ShortUrl:      shortenURL,
			},
		)
	}
	err = srv.s.AddURLBatch(ctx, urlIDs, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &pb.GetShortURLBatchResponse{Batch: result}, nil
}

// DeleteURL is a method to delete URLs.
func (srv *ShortyServer) DeleteURL(stream pb.Shorty_DeleteURLServer) error {
	srv.m.Lock()
	defer srv.m.Unlock()
	userID, err := getUserID(stream.Context())
	if err != nil {
		return err
	}

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.DeleteURLResponse{})
		}
		if err != nil {
			return err
		}
		if _, ok := srv.deleteJobs[userID]; !ok {
			srv.deleteJobs[userID] = make([]string, 0)
		}
		srv.deleteJobs[userID] = append(srv.deleteJobs[userID], req.Url)
	}
}

// deleteURLLoop is a method to organize URLs deleting.
func (srv *ShortyServer) deleteURLLoop() {
	ticker := time.NewTicker(srv.expireDuration)
out:
	for {
		select {
		case <-ticker.C:
			srv.m.Lock()
			for userID, userJobs := range srv.deleteJobs {
				err := srv.s.DeleteMany(context.Background(), userID, userJobs)
				if err != nil {
					log.Printf("Error while deleting urls: %v\n", err)
				}
			}
			srv.deleteJobs = make(map[uint32][]string)
			srv.m.Unlock()
		case <-srv.srvCloseCh:
			break out
		}
	}
	srv.m.Lock()
	for userID, userJobs := range srv.deleteJobs {
		err := srv.s.DeleteMany(context.Background(), userID, userJobs)
		if err != nil {
			log.Printf("Error while deleting urls: %v\n", err)
		}
	}
	srv.m.Unlock()
	srv.srvCloseCh <- struct{}{}
}

// GetStats is a method to retrieve DB stats.
func (srv *ShortyServer) GetStats(
	ctx context.Context,
	req *pb.GetStatsRequest,
) (*pb.GetStatsResponse, error) {
	if srv.trustedSubnet == nil {
		return nil, status.Errorf(codes.Internal, "trusted network is not set")
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "No metadata provided")
	}
	ips := md.Get("X-Real-IP")
	var ipStr string
	if len(ips) > 0 {
		ipStr = ips[0]
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, status.Errorf(codes.Internal, "failed parse ip from metadata")
	}
	if !srv.trustedSubnet.Contains(ip) {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("ip %v is not in trusted network %+v", ip, srv.trustedSubnet),
		)
	}
	urls, users, err := srv.s.GetStats(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &pb.GetStatsResponse{Users: uint32(users), Urls: uint32(urls)}, nil

}

// Ping is a method for checking DB status.
func (srv *ShortyServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{Pinged: srv.s.Ping(ctx)}, nil
}
