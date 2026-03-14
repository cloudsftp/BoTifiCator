package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"resty.dev/v3"

	"github.com/go-co-op/gocron/v2"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"github.com/cloudsftp/botificator/pkg/analyzer"
	"github.com/cloudsftp/botificator/pkg/db"
	"github.com/cloudsftp/botificator/pkg/load"
	"github.com/cloudsftp/botificator/pkg/notificator"
)

type Server struct {
	notificator  *notificator.Notificator
	dataProvider *db.DataProvider
	client       *resty.Client
	scheduler    gocron.Scheduler
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

	dataProvider, err := db.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not setup database: %w", err)
	}

	client := resty.New()

	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("could not create scheduler: %w", err)
	}

	return &Server{notificator, dataProvider, client, scheduler}, nil
}

func (s *Server) Close() {
	err := s.client.Close()
	if err != nil {
		logrus.Errorf("could not close client: %s", err)
	}

	s.dataProvider.Close()
}

func (s *Server) Run(ctx context.Context) error {
	err := s.notificator.SendMessageDeployed(ctx)
	if err != nil {
		return fmt.Errorf("could not send initial message: %w", err)
	}

	var done sync.WaitGroup
	done.Add(1)
	errors := make(chan error)
	go func() {
	loop:
		for {
			select {
			case err := <-errors:
				logrus.Errorf("runtime error: %s", err)
			case <-ctx.Done(): // TODO: remove when other cases are added
				logrus.Error("context done")
				break loop
			}
		}
		done.Done()
	}()

	err = s.UpdateDatabase(ctx)
	if err != nil {
		return fmt.Errorf("could not update database initially: %w", err)
	}

	err = s.SendUpdate(ctx)
	if err != nil {
		return fmt.Errorf("could not send initial update message: %w", err)
	}

	_, err = s.scheduler.NewJob(
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

	done.Wait()
	return nil
}

func (s *Server) UpdateDatabase(ctx context.Context) error {
	err := load.LoadDataIntoDatabase(ctx, s.client, s.dataProvider)
	if err != nil {
		return fmt.Errorf("could not load data into database: %s", err)
	}

	logrus.Debug("Updated database")
	return nil
}

func (s *Server) SendUpdate(ctx context.Context) error {
	report, err := analyzer.Analyze(ctx, s.dataProvider)
	if err != nil {
		return fmt.Errorf("could not generate daily reports: %s", err)
	}

	logrus.Debug("generated reports successfully")

	err = s.notificator.SendDailyReports(ctx, report)
	if err != nil {
		return fmt.Errorf("could not send daily reports: %s", err)
	}

	return nil
}
