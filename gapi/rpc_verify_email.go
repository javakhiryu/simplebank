package gapi

import (
	"context"
	db "simplebank/db/sqlc"
	"simplebank/pb"
	"simplebank/valid"
	_ "simplebank/valid"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	violations := validateVerifyEmailRequest(req)
	if violations != nil {
		return nil, violationsError(violations)
	}

	txResult, err := server.store.VerifyEmailTx(ctx, db.VerifyEmailTxParams{
		EmailId:    req.GetEmailId(),
		SecretCode: req.GetSecretCode()})
	if err != nil {
		if err == db.ErrNoRowsFound {
			return nil, status.Errorf(codes.NotFound, "verification code not found: %s", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to verify email: %s", err)
	}

	rsp := &pb.VerifyEmailResponse{
		IsEmailVerified: txResult.User.IsEmailVerified,
	}

	return rsp, nil
}

func validateVerifyEmailRequest(req *pb.VerifyEmailRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := valid.ValidateVerifyEmailId(req.GetEmailId()); err != nil {
		violations = append(violations, fieldViolation("email_id", err))
	}
	if err := valid.ValidateSecretCode(req.GetSecretCode()); err != nil {
		violations = append(violations, fieldViolation("secret_code", err))
	}
	return violations
}
