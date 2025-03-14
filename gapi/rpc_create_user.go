package gapi

import (
	"context"
	db "simplebank/db/sqlc"
	"simplebank/pb"
	"simplebank/util"
	"simplebank/valid"
	_ "simplebank/valid"
	"simplebank/worker"
	"time"

	"github.com/hibiken/asynq"
	"github.com/lib/pq"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	violations := validateCreateUserRequest(req)
	if violations != nil {
		return nil, violationsError(violations)
	}
	hashedPassword, err := util.HashedPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
	}

	arg := db.CreateUserTxParams{
		CreateUserParams: db.CreateUserParams{
			Username: req.GetUsername(),
			HashedPassword: hashedPassword,
			FullName: req.GetFullName(),
			Email:    req.GetEmail(),
	},
	AfterCreate: func(user db.User) error {
		taskPayload := worker.PayloadSendVerifyEmail{
			Username: user.Username,
		}
		options := []asynq.Option{
			asynq.MaxRetry(10),
			asynq.ProcessIn(10 * time.Second),
			asynq.Queue(worker.CriticalQueue),
		}
	
		return server.taskDistributor.DistributeTaskSendVerifyEmail(ctx, &taskPayload, options...)
	},
}

	txResult, err := server.store.CreateUserTx(ctx, arg)
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code.Name() {
		case "unique_violation":
			return nil, status.Errorf(codes.AlreadyExists, "username already exists: %v", err)
		}
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}
	
	rsp := &pb.CreateUserResponse{
		User: convertUser(txResult.User),
	}

	return rsp, nil
}

func validateCreateUserRequest(req *pb.CreateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := valid.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}
	if err := valid.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}
	if err := valid.ValidateEmail(req.GetEmail()); err != nil {
		violations = append(violations, fieldViolation("email", err))
	}
	if err := valid.ValidateFullName(req.GetFullName()); err != nil {
		violations = append(violations, fieldViolation("full_name", err))
	}
	return violations
}
