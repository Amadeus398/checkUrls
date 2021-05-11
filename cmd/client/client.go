package client

import (
	"CheckUrls/pkg/logging"
	"CheckUrls/pkg/proto"
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"strconv"
)

var IncorrectInput = fmt.Errorf("incorrect input of arguments")

type CliConfig interface {
	GetServerAddress() string
}

func StartClient(cfg CliConfig) (proto.SitesServiceClient, error) {
	logger := logging.NewLoggers("client", "startClient")
	logger.DebugLog().Msg("connecting to server")
	conn, err := grpc.Dial(cfg.GetServerAddress(), grpc.WithInsecure())
	if err != nil {
		logger.ErrorLog().Str("when", "dial :8000").Err(err).Msg("failed start client")
		return nil, err
	}

	return proto.NewSitesServiceClient(conn), nil
}

func ReqCreateSite(ctx context.Context, cli proto.SitesServiceClient) error {
	logger := logging.NewLoggers("client", "reqCreate")
	logger.DebugLog().Msg("checking for the correctness of arguments")
	if flag.NArg() < 3 {
		err := IncorrectInput
		logger.ErrorLog().Err(err).Str("request", "failed to process").
			Msg("please enter \"create <url> <frequency>\"")
		return err
	}

	logger.DebugLog().Msg("getting arguments")
	url := flag.Arg(2)
	frequency := 0
	var err error
	if flag.Arg(3) != "" {
		frequency, err = strconv.Atoi(flag.Arg(3))
		if err != nil {
			logger.ErrorLog().Err(err).Str("request", "failed to process").
				Msg("cannot to convert frequency")
			return err
		}
	}

	logger.DebugLog().Msg("create site")
	res, err := cli.Create(ctx, &proto.CreateRequestSite{
		Sites: &proto.Site{
			Url:       url,
			Frequency: int64(frequency),
		},
	})
	if err != nil {
		logger.ErrorLog().Err(err).Str("request", "failed to process").
			Msg("unable to create site")
		return err
	}
	logger.InfoLog().Str("request", "processed successfully").
		Interface("a new site was created with site_id: ", res.GetId()).Msg("done")

	return nil
}

func ReqReadSite(ctx context.Context, cli proto.SitesServiceClient) error {
	logger := logging.NewLoggers("client", "reqRead")
	logger.DebugLog().Msg("checking for the correctness of arguments")
	if flag.NArg() != 3 {
		err := IncorrectInput
		logger.ErrorLog().Err(err).Str("request", "failed to process").
			Msg("please enter \"read <site_id>\"")
		return err
	}

	logger.DebugLog().Msg("getting arguments")
	id, err := strconv.Atoi(flag.Arg(2))
	if err != nil {
		logger.ErrorLog().Err(err).Str("request", "failed to process").
			Msg("cannot to convert site_id")
		return err
	}

	logger.DebugLog().Msg("getting site")
	res, err := cli.Read(ctx, &proto.ReadRequestSite{Id: int64(id)})
	if err != nil {
		logger.ErrorLog().Err(err).Str("request", "failed to process").
			Msg("unable to get site")
		return err
	}
	logger.InfoLog().Str("request", "processed successfully").
		Interface("read result: ", res.GetSites()).Msg("done")

	return nil
}

func ReqReadAllSite(ctx context.Context, cli proto.SitesServiceClient) error {
	logger := logging.NewLoggers("client", "reqReadAllSite")
	logger.DebugLog().Msg("checking for the correctness of arguments")
	if flag.NArg() != 2 {
		err := IncorrectInput
		logger.ErrorLog().Err(err).Str("request", "failed to process").
			Msg("please enter \"list\"")
		return err
	}

	logger.DebugLog().Msg("getting list of sites")
	res, err := cli.ReadAll(ctx, &proto.ReadAllRequestSite{})
	if err != nil {
		logger.ErrorLog().Err(err).Str("request", "failed to process").
			Msg("unable to get list of sites")
		return err
	}
	logger.InfoLog().Str("request", "processed successfully").
		Interface("list of sites: ", res.GetSites()).Msg("done")

	return nil
}

func ReqUpdateSite(ctx context.Context, cli proto.SitesServiceClient) error {
	logger := logging.NewLoggers("client", "reqUpdate")
	logger.DebugLog().Msg("checking for the correctness of arguments")
	if flag.NArg() < 5 {
		err := IncorrectInput
		logger.ErrorLog().Err(err).Str("request", "failed to process").
			Msg("please enter \"update <site_id> <url> <frequency>\"")
		return err
	}

	logger.DebugLog().Msg("getting arguments")
	id, err := strconv.Atoi(flag.Arg(2))
	if err != nil {
		logger.ErrorLog().Err(err).Str("request", "failed to process").
			Msg("cannot to convert site_id")
		return err
	}
	url := flag.Arg(3)
	frequency, err := strconv.Atoi(flag.Arg(4))
	if err != nil {
		logger.ErrorLog().Err(err).Str("request", "failed to process").
			Msg("cannot to convert site_id")
		return err
	}
	site := &proto.Site{
		Id:        int64(id),
		Url:       url,
		Frequency: int64(frequency),
	}

	logger.DebugLog().Msg("updating site")
	res, err := cli.Update(ctx, &proto.UpdateRequestSite{Sites: site})
	if err != nil {
		logger.ErrorLog().Err(err).Str("request", "failed to process").
			Msg("unable to update site")
		return err
	}
	logger.InfoLog().Str("request", "processed successfully").
		Interface("updated: ", res.GetUpdated()).Msg("done")

	return nil
}

func ReqDeleteSite(ctx context.Context, cli proto.SitesServiceClient) error {
	logger := logging.NewLoggers("client", "reqDelete")
	logger.DebugLog().Msg("checking for the correctness of arguments")
	if flag.NArg() != 3 {
		err := IncorrectInput
		logger.WarnLog().Err(err).Str("request", "failed to process").
			Msg("please enter \"delete <site_id>\"")
		return err
	}

	logger.DebugLog().Msg("getting arguments")
	id, err := strconv.Atoi(flag.Arg(2))
	if err != nil {
		logger.WarnLog().Err(err).Str("request", "failed to process").
			Msg("cannot to convert site_id")
		return err
	}

	logger.DebugLog().Msg("deleting site")
	res, err := cli.Delete(ctx, &proto.DeleteRequestSite{Id: int64(id)})
	if err != nil {
		logger.ErrorLog().Err(err).Str("request", "failed to process").
			Msg("unable to delete site")
		return err
	}
	logger.InfoLog().Str("request", "processed successfully").
		Interface("deleted: ", res.GetDeleted())

	return nil
}

func ReqReadStatus(ctx context.Context, cli proto.SitesServiceClient) error {
	logger := logging.NewLoggers("client", "reqReadStatus")
	logger.DebugLog().Msg("checking for the correctness of arguments")
	if flag.NArg() < 3 {
		err := IncorrectInput
		logger.ErrorLog().Err(err).Str("request", "failed to process").
			Msg("please enter \"<url>\"")
		return err
	}

	logger.DebugLog().Msg("getting arguments")
	url := flag.Arg(2)
	count := 5
	var err error
	if flag.Arg(3) != "" {
		count, err = strconv.Atoi(flag.Arg(3))
		if err != nil {
			logger.ErrorLog().Err(err).Str("when", "convert count").Msg("unable to convert count")
			return err
		}
	}

	logger.DebugLog().Msg("read request processing")
	res, err := cli.ReadStatus(ctx, &proto.ReadRequestState{
		Url:   url,
		Count: int64(count),
	})
	if err != nil {
		logger.ErrorLog().Err(err).Str("request", "failed to process").
			Msg("unable to get list of states")
		return err
	}
	logger.InfoLog().Str("request", "processed successfully").
		Interface("list of states: ", res.GetStates()).Msg("done")

	return nil
}
