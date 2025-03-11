package gapi

import (
	"context"
	"database/sql"
	db "simplebank/db/sqlc"
	"simplebank/pb"
	"simplebank/util"
	"simplebank/valid"
	_ "simplebank/valid"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) UpdateUserPassword(ctx context.Context, req *pb.UpdateUserPasswordRequest) (*pb.UpdateUserPasswordResponse, error) {
	authPayload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unauthencatedError(err)
	}
	violations := validateUpdateUserPasswordRequest(req)
	if violations != nil {
		return nil, violationsError(violations)
	}

	if authPayload.Username != req.GetUsername() {
		return nil, status.Errorf(codes.PermissionDenied, "cannot update other user's information")
	}

	user, err := server.store.GetUser(ctx, authPayload.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	err = util.CheckPassword(req.GetOldPassword(), user.HashedPassword)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "old password is incorrect: %v", err)
	}

	hashedPassword, err := util.HashedPassword(req.GetNewPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
	}
	arg := db.UpdateUserHashedPasswordParams{
		Username:       req.GetUsername(),
		HashedPassword: hashedPassword,
	}

	user, err = server.store.UpdateUserHashedPassword(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to update user password: %v", err)
	}
	rsp := &pb.UpdateUserPasswordResponse{
		User: convertUser(user),
	}

	return rsp, nil
}

func validateUpdateUserPasswordRequest(req *pb.UpdateUserPasswordRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := valid.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}
	if err := valid.ValidatePassword(req.GetOldPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}
	if err := valid.ValidatePassword(req.GetNewPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}
	return violations
}
