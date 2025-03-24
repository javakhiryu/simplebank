package gapi

import (
	"context"
	"fmt"
	"simplebank/token"
	"strings"

	"google.golang.org/grpc/metadata"
)

const (
	authorizationHeaderKey = "authorization"
	authorizationBearer    = "bearer"
)

func (Server *Server) authorizeUser(ctx context.Context, accessibleRole []string) (*token.Payload, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}
	values := md.Get(authorizationHeaderKey)
	if len(values) == 0 {
		return nil, fmt.Errorf("missing authorization header")
	}

	authHeader := values[0]
	fields := strings.Fields(authHeader)
	if len(fields) < 2 {
		return nil, fmt.Errorf("invalid authrization header format")
	}

	authType := strings.ToLower(fields[0])
	if authType != authorizationBearer {
		return nil, fmt.Errorf("unsupported authorization type %s", authType)
	}

	accessToken := fields[1]
	payload, err := Server.tokenMaker.VerifyToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("invalid access token: %w", err)
	}

	if !hasPermission(payload.Role, accessibleRole) {
		return nil, fmt.Errorf("permission denied")
	}

	return payload, nil
}

func hasPermission(userRole string, accessibleRole []string) bool{
	for _, role := range accessibleRole {
		if userRole == role{
			return true
		}
	}
	return false
}
