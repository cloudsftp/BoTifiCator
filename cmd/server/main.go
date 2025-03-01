package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"resty.dev/v3"

	"github.com/go-co-op/gocron/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/cloudsftp/botificator/pkg/analyzer"
	"github.com/cloudsftp/botificator/pkg/db"
	"github.com/cloudsftp/botificator/pkg/load"
	"github.com/cloudsftp/botificator/pkg/notificator"
)

var startTime = time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

func main() {
	ctx := context.Background()

	server, err := NewServer(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error in setup: %s", err)
		os.Exit(1)
	}
	defer server.Close()

	err = server.Run(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while running: %s", err)
		os.Exit(1)
	}
}

type Server struct {
	notificator  *notificator.Notificator
	pool         *pgxpool.Pool
	client       *resty.Client
	scheduler    gocron.Scheduler
	errors       chan error
	databaseLock *sync.RWMutex
}

func NewServer(ctx context.Context) (*Server, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("could not load environment: %w\n", err)
	}

	notificator, err := notificator.New()
	if err != nil {
		return nil, fmt.Errorf("could not create notificator: %w\n", err)
	}

	pool, err := db.SetupDatabase(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not setup database: %w\n", err)
	}

	client := resty.New()

	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("could not create scheduler: %w\n", err)
	}

	err = notificator.SendMessageDeployed(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not send message: %w\n", err)
	}

	errors := make(chan error)
	var databaseLock *sync.RWMutex

	return &Server{notificator, pool, client, scheduler, errors, databaseLock}, nil
}

func (s *Server) Close() {
	s.pool.Close()
	s.client.Close()
}

func (s *Server) Run(ctx context.Context) error {
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

	for {
		select {
		case err := <-s.errors:
			fmt.Fprintf(os.Stderr, "runtime error: %s", err)
		}
	}
}

func (s *Server) UpdateDatabase(ctx context.Context) {
	s.databaseLock.Lock()
	defer s.databaseLock.Unlock()

	err := load.LoadDataIntoDatabase(ctx, s.client, s.pool, startTime)
	if err != nil {
		s.errors <- fmt.Errorf("could not load data into database: %s\n", err)
	}
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
		s.errors <- fmt.Errorf("could not get moving averages: %s\n", err)
	}

	err = analyzer.Analyze(averages)
	if err != nil {
		s.errors <- fmt.Errorf("could not analyze averages: %s\n", err)
	}
}
