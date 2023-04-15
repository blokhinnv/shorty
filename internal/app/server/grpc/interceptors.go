package grpc

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/blokhinnv/shorty/internal/app/server/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UserTokenMDName is a metadata key containing user token.
const UserTokenMDName = "UserToken"

// getUserIDInteceptor returns uid or creates one from random user token.
func getUserIDInteceptor(md metadata.MD, secretKey []byte) (uint32, error) {
	// get userID (or generate new)
	userTokenMD, ok := md[strings.ToLower(UserTokenMDName)]
	var userTokenStr string
	var err error
	if !ok {
		userTokenStr, err = auth.GenerateToken(secretKey)
		if err != nil {
			return 0, status.Errorf(codes.Internal, "Unable to generate token")
		}
	} else {
		userTokenStr = userTokenMD[0]
	}
	userToken, err := hex.DecodeString(userTokenStr)
	if err != nil {
		return 0, status.Errorf(codes.Internal, "Unable to decode token")
	}
	userID := auth.ExtractID(userToken)
	return userID, nil
}

// userTokenInceptor is a server interceptor which updates the metadata with UID.
func (srv *ShortyServer) userTokenInceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	// src: https://shijuvar.medium.com/writing-grpc-interceptors-in-go-bf3e7671fe48
	// retrieve metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Errorf(codes.InvalidArgument, "Retrieving metadata is failed")
	}
	userID, err := getUserIDInteceptor(md, srv.secretKey)
	if err != nil {
		return nil, err
	}
	// update metadata
	md.Set(UserTokenMDName, fmt.Sprint(userID))
	newCtx := metadata.NewIncomingContext(ctx, md)
	h, err := handler(newCtx, req)
	return h, err
}

// userTokenServerStream is a type to implement metadata update for a stream interceptor.
type userTokenServerStream struct {
	grpc.ServerStream
	userID uint32
}

// Context is an implementation of userTokenServerStream.Context.
func (u *userTokenServerStream) Context() context.Context {
	ctx := u.ServerStream.Context()
	md := metadata.Pairs("UserToken", fmt.Sprint(u.userID))
	return metadata.NewIncomingContext(ctx, md)
}

// userTokenStreamInterceptor is a server stream interceptor which updates the metadata with UID.
func (srv *ShortyServer) userTokenStreamInterceptor(
	srvImpl any,
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	// retrieve metadata
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return status.Errorf(codes.InvalidArgument, "Retrieving metadata is failed")
	}
	userID, err := getUserIDInteceptor(md, srv.secretKey)
	if err != nil {
		return err
	}
	err = handler(srv, &userTokenServerStream{stream, userID})
	if err != nil {
		return err
	}
	return nil
}
