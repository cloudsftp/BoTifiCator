package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudsftp/botificator/pkg/analyzer"
	"github.com/cloudsftp/botificator/pkg/db"
	"github.com/cloudsftp/botificator/pkg/load"
	"github.com/cloudsftp/botificator/pkg/notificator"
	"github.com/go-co-op/gocron/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"resty.dev/v3"
)

type Server struct {
	notificator  *notificator.Notificator
	pool         *pgxpool.Pool
	client       *resty.Client
	scheduler    gocron.Scheduler
	errors       chan error
	databaseLock *sync.RWMutex
}

func New(ctx context.Context) (*Server, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("could not load environment: %w", err)
	}

	notificator, err := notificator.New()
	if err != nil {
		return nil, fmt.Errorf("could not create notificator: %w", err)
	}

	pool, err := db.SetupDatabase(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not setup database: %w", err)
	}

	client := resty.New()

	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("could not create scheduler: %w", err)
	}

	err = notificator.SendMessageDeployed(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not send message: %w", err)
	}

	errors := make(chan error)
	var databaseLock sync.RWMutex

	return &Server{notificator, pool, client, scheduler, errors, &databaseLock}, nil
}

func (s *Server) Close() {
	s.pool.Close()
	s.client.Close()
}

func (s *Server) Run(ctx context.Context) error {
	s.UpdateDatabase(ctx)
	_, err := s.scheduler.NewJob(
		gocron.DurationJob(15*time.Minute),
		gocron.NewTask(s.UpdateDatabase, ctx),
	)
	if err != nil {
		return fmt.Errorf("could not set up database cron job: %w", err)
	}

	_, err = s.scheduler.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(5, 0, 0))),
		gocron.NewTask(s.SendUpdate, ctx),
	)
	if err != nil {
		return fmt.Errorf("could not set up notification cron job: %w", err)
	}

	s.scheduler.Start()
	defer func() {
		err := s.scheduler.Shutdown()
		if err != nil {
			logrus.Errorf("could not shutdown scheduler: %s", err)
		}
	}()

	for {
		select {
		case err := <-s.errors:
			logrus.Errorf("runtime error: %s", err)
		}
	}
}

func (s *Server) UpdateDatabase(ctx context.Context) {
	ok := s.databaseLock.TryLock()
	if !ok {
		// TODO: warn being updated already
		return
	}
	defer s.databaseLock.Unlock()

	err := load.LoadDataIntoDatabase(ctx, s.client, s.pool)
	if err != nil {
		s.errors <- fmt.Errorf("could not load data into database: %s", err)
	}

	logrus.Debug("Updated database")
}

func (s *Server) SendUpdate(ctx context.Context) {
	ok := s.databaseLock.TryRLock()
	if !ok {
		// TODO: wait for an hour, otherwise error out
		return
	}
	defer s.databaseLock.RUnlock()

	averages, err := db.GetMovingAverages(ctx, s.pool)
	if err != nil {
		s.errors <- fmt.Errorf("could not get moving averages: %s", err)
	}

	err = analyzer.Analyze(averages)
	if err != nil {
		s.errors <- fmt.Errorf("could not analyze averages: %s", err)
	}
}
