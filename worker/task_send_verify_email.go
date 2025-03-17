package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	db "simplebank/db/sqlc"
	"simplebank/util"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const TaskSendVerifyEmail = "task:send_verify_email"

type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
}

func (distributor *RedisTaskDistributor) DistributeTaskSendVerifyEmail(
	ctx context.Context,
	payload *PayloadSendVerifyEmail,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}
	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}
	log.Info().
		Str("type:", info.Type).
		Bytes("payload:", info.Payload).
		Str("queue:", info.Queue).
		Int("max_retry", info.MaxRetry).
		Msg("enqueued task")
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendVerifyEmail
	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}
	user, err := processor.store.GetUser(ctx, payload.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user does not exist: %w", asynq.SkipRetry)
		}
		return fmt.Errorf("failed to fetch user: %w", err)
	}
	secretCode := util.RandomString(32)
	Id := uuid.New()

	verifyEmail, err := processor.store.CreateVerifyEmail(ctx, db.CreateVerifyEmailParams{
		ID:         Id,
		Username:   user.Username,
		Email:      user.Email,
		SecretCode: secretCode,
	})
	if err != nil {
		return fmt.Errorf("failed to create verify email record: %w", err)
	}
	subject := "Welcome to Simple Bank"
	verifyLink := fmt.Sprintf("http://localhost:8080?id=%s&secret_code=%s", verifyEmail.ID, verifyEmail.SecretCode)
	content := fmt.Sprintf(`
		<html>
		<head>
			<title>Email Verification</title>
		</head>
		<body>
			<h1>Dear %s,</h1>
			<p>Thank you for registering with us! Please click the link below to verify your email address:</p>
			<p><a href="%s">Verify Email</a></p>
			<p>If you did not create an account, please ignore this email.</p>
		</body>
		</html>
	`, user.Username, verifyLink)
	to := []string{user.Email}
	err = processor.mailer.SendEmail(
		subject,
		content,
		to,
		nil,
		nil,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to send verify email: %w", err)
	}
	log.Info().
		Str("type:", task.Type()).
		Bytes("payload: ", task.Payload()).
		Str("email: ", user.Email).
		Msg("task processed successfully")

	return nil
}
