package grpc

import (
	"context"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// getUserID return uint32 user id from metadata after the interceptor has updated it.
func getUserID(ctx context.Context) (uint32, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Errorf(codes.PermissionDenied, "No metadata provided")
	}
	var userIDStr string

	userIDs := md.Get(strings.ToLower(UserTokenMDName))
	if len(userIDs) > 0 {
		userIDStr = userIDs[0]
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return 0, status.Errorf(codes.InvalidArgument, "Incorrect userid provided")
	}
	return uint32(userID), nil
}
