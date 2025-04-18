package worker

import (
	"context"
	db "simplebank/db/sqlc"
	"simplebank/mail"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	CriticalQueue = "critical"
	DefaultQueue  = "default"
	LowQueue      = "low"
)

type TaskProcessor interface {
	ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error
	Start() error
	Shutdown()
}

type RedisTaskProcessor struct {
	server *asynq.Server
	store  db.Store
	mailer mail.EmailSender
}

func NewRedisTaskProcessor(mailer mail.EmailSender, redisOtp asynq.RedisClientOpt, store db.Store) TaskProcessor {
	redis.SetLogger(NewLogger())
	server := asynq.NewServer(redisOtp, asynq.Config{
		Queues: map[string]int{
			CriticalQueue: 6,
			DefaultQueue:  3,
			LowQueue:      1,
		},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			log.Error().Err(err).Msg("process task failed")
		}),
		Logger: NewLogger(),
	},
	)
	return &RedisTaskProcessor{
		server: server,
		store:  store,
		mailer: mailer,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskSendVerifyEmail, processor.ProcessTaskSendVerifyEmail)
	return processor.server.Start(mux)
}

func (processor *RedisTaskProcessor) Shutdown(){
	processor.server.Shutdown()
}
