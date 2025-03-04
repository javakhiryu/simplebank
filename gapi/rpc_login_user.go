package gapi

import (
	"context"
	"database/sql"
	db "simplebank/db/sqlc"
	"simplebank/pb"
	"simplebank/util"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {

	user, err := server.store.GetUser(ctx, req.GetUsername())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	err = util.CheckPassword(req.GetPassword(), user.HashedPassword)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "password is incorrect: %v", err)
	}
	accessToken, accessPayload, err := server.tokenMaker.CreateToken(req.GetUsername(), server.config.AccessTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create access token: %v", err)
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(req.Username, server.config.RefreshTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create refresh token: %v", err)
	}
	metadata := server.extractMetadata(ctx)
	arg := db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     req.GetUsername(),
		RefreshToken: refreshToken,
		UserAgent:    metadata.UserAgent,
		IpAddress:    metadata.IpAddress,
		ExpiresAt:    refreshPayload.ExpiredAt,
		CreatedAt:    time.Now(),
		IsBlocked:    false,
	}
	session, err := server.store.CreateSession(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create session: %v", err)
	}

	rsp := &pb.LoginUserResponse{
		SessionId:             session.ID.String(),
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		User:                  convertUser(user),
		AccessTokenExpiresAt:  timestamppb.New(accessPayload.ExpiredAt),
		RefreshTokenExpiresAt: timestamppb.New(refreshPayload.ExpiredAt),
	}
	return rsp, nil
}
