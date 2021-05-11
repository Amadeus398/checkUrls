package main

import (
	"CheckUrls/cmd/client"
	server "CheckUrls/cmd/server"
	"CheckUrls/pkg/backendMngr"
	"CheckUrls/pkg/config"
	"CheckUrls/pkg/db"
	"CheckUrls/pkg/logging"
	"context"
	"flag"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"os"
	"os/signal"
)

type loggerConfig interface {
	GetLogLevel() zerolog.Level
}

func main() {
	logger := logging.NewLoggers("cmd", "main")

	var correctExit = fmt.Errorf("finished successfully")
	cfg := &config.EnvCache{}
	if err := envconfig.Process("", cfg); err != nil {
		logger.FatalLog().Str("when", "parse environment").Err(err).
			Msg("unable to parse the environment")
	}

	logCfg := loggerConfig(cfg)

	zerolog.SetGlobalLevel(logCfg.GetLogLevel())

	ctx := context.TODO()
	flag.Parse()
	switch flag.Arg(0) {
	case "server":
		dbCfg := db.DbConfig(cfg)
		serveCfg := server.ServerConfig(cfg)

		connMnr := db.NewConnectionManager()

		if err := connMnr.Connect(dbCfg); err != nil {
			logger.FatalLog().Str("when", "connect DB").Err(err).Msg("failed to connect DB")
		}
		defer func() {
			if err := connMnr.Close(); err != nil {
				logger.FatalLog().Str("when", "close connect DB").Err(err).Msg("failed to close DB")
			}
		}()
		logger.InfoLog().Str("when", "starting server").Msg("connecting DB")

		errGroup, errGroupCtx := errgroup.WithContext(ctx)
		s := grpc.NewServer()
		serve := &server.GRPCServer{
			Backend: backendMngr.NewBackendManager(connMnr, errGroupCtx),
			Сonn:    connMnr,
		}

		errGroup.Go(func() error {
			interruptChan := make(chan os.Signal, 1)
			signal.Notify(interruptChan, os.Interrupt)
			select {
			case <-interruptChan:
				logger.InfoLog().Str("when", "got shutdown signal").
					Msg("got signal interrupt attempting graceful shutdown")
				s.GracefulStop()
				return correctExit
			case <-errGroupCtx.Done():
				logger.InfoLog().Str("when", "got context signal").Msg("received context closure signal")
				s.GracefulStop()
				return errGroupCtx.Err()
			}
		})

		errGroup.Go(func() error {
			err := server.RunServer(serveCfg, errGroupCtx, serve, s)
			if err != nil {
				logger.ErrorLog().Str("when", "starting server").Err(err).
					Msg("failed to start grpc server")
				return err
			}
			return nil
		})
		logger.InfoLog().Str("when", "start server").Msg("server is listening...")

		if err := errGroup.Wait(); err != nil {
			if err == correctExit {
				return
			}
			logger.ErrorLog().Str("when", "have error from server").Err(err).Msg("exiting")
			os.Exit(1)
		}

	case "client":
		cliCfg := client.CliConfig(cfg)
		cli, err := client.StartClient(cliCfg)
		if err != nil {
			logger.FatalLog().Str("when", "starting client").Err(err).Msg("failed to start client")
		}
		switch flag.Arg(1) {
		case "create":
			logger.InfoLog().Str("when", "start client").Msg("creating site")
			if err := client.ReqCreateSite(ctx, cli); err != nil {
				logger.FatalLog().Str("when", "create site").Err(err).Msg("failed to create site")
			}
		case "read":
			logger.InfoLog().Str("when", "start client").Msg("getting site")
			if err := client.ReqReadSite(ctx, cli); err != nil {
				logger.FatalLog().Str("when", "get site").Err(err).Msg("failed to get site")
			}
		case "list":
			if err := client.ReqReadAllSite(ctx, cli); err != nil {
				logger.InfoLog().Str("when", "start client").Msg("getting list of sites")
				logger.FatalLog().Str("when", "get list of sites").Err(err).
					Msg("failed to get list of sites")
			}
		case "update":
			logger.InfoLog().Str("when", "start client").Msg("updating site")
			if err := client.ReqUpdateSite(ctx, cli); err != nil {
				logger.FatalLog().Str("when", "update site").Err(err).Msg("failed to update site")
			}
		case "delete":
			logger.InfoLog().Str("when", "start client").Msg("deleting site")
			if err := client.ReqDeleteSite(ctx, cli); err != nil {
				logger.FatalLog().Str("when", "delete site").Err(err).Msg("failed to delete site")
			}
		case "status":
			logger.InfoLog().Str("when", "start client").Msg("getting list of statuses")
			if err := client.ReqReadStatus(ctx, cli); err != nil {
				logger.FatalLog().Str("when", "get list of statuses").Err(err).
					Msg("failed to get list of statuses")
			}
		default:
			err := client.IncorrectInput
			logger.FatalLog().Str("when", "entering a sites request").Err(err).
				Msg("please enter operation (create, read, update, delete or status)")
		}
	default:
		err := client.IncorrectInput
		logger.FatalLog().Str("when", "starting server").Err(err).
			Msg("enter \"server\" or \"client\" operation")
	}
}
