package server

import (
	"CheckUrls/pkg/backendMngr"
	"CheckUrls/pkg/logging"
	"CheckUrls/pkg/proto"
	"CheckUrls/pkg/repository/sites"
	statuses "CheckUrls/pkg/repository/status"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
)

// TODO configs
type ServerConfig interface {
	GetServerAddress() string
}

// GRPCServer ...
type GRPCServer struct {
	proto.UnimplementedSitesServiceServer
	Backend *backendMngr.BackendManager
	log     *logging.Loggers
}

// Create site...
func (g *GRPCServer) Create(ctx context.Context, request *proto.CreateRequestSite) (*proto.CreateResponseSite, error) {
	g.log = logging.NewLoggers("server", "create")
	site := sites.Site{
		Url:       request.GetSites().Url,
		Frequency: request.GetSites().Frequency,
	}
	if err := sites.CreateSites(&site); err != nil {
		if err == sites.ErrSitesNotFound {
			err = status.Error(codes.NotFound, "unable to create site")
			g.log.WarnLog().Str("when", "create site").Str("request", "failed to process").
				Err(err).Msg("unable to create site")
		} else {
			err = status.Error(codes.Unknown, "unable to create site")
			g.log.ErrorLog().Str("when", "create site").Str("request", "failed to process").
				Err(err).Msg("unable to create site")
		}
		return nil, err
	}

	g.Backend.CreateOrUpdate(&site)
	return &proto.CreateResponseSite{Id: site.Id}, nil
}

// Read site...
func (g *GRPCServer) Read(ctx context.Context, request *proto.ReadRequestSite) (*proto.ReadResponseSite, error) {
	g.log = logging.NewLoggers("server", "read")
	site := sites.Site{Id: request.GetId()}
	if err := sites.ReadSites(&site); err != nil {
		if err == sites.ErrSitesNotFound {
			err = status.Error(codes.NotFound, "unable to get site")
			g.log.WarnLog().Str("when", "get site").Str("request", "failed to process").
				Err(err).Msg("unable to get site")
		} else {
			err = status.Error(codes.Unknown, "unable to get site")
			g.log.ErrorLog().Str("when", "get site").Str("request", "failed to process").
				Err(err).Msg("unable to get site")
		}
		return nil, err
	}

	siteProto := proto.Site{
		Id:        site.Id,
		Url:       site.Url,
		Frequency: site.Frequency,
	}

	return &proto.ReadResponseSite{
		Sites: &siteProto,
	}, nil
}

// ReadAll is list of sites...
func (g *GRPCServer) ReadAll(ctx context.Context, request *proto.ReadAllRequestSite) (*proto.ReadAllResponseSite, error) {
	g.log = logging.NewLoggers("server", "readAll")
	list, err := sites.ReadAllSites()
	if err != nil {
		if err == sites.ErrSitesNotFound {
			err = status.Error(codes.NotFound, "unable to get list")
			g.log.WarnLog().Str("when", "get list of sites").Str("request", "failed to process").
				Err(err).Msg("unable to get list of sites")
		} else {
			err = status.Error(codes.Unknown, "unable to get list")
			g.log.WarnLog().Str("when", "get list of sites").Str("request", "failed to process").
				Err(err).Msg("unable to get list of sites")
		}
		return nil, err
	}

	listProto := make([]*proto.Site, 0, len(list))
	for _, site := range list {
		ls := new(proto.Site)
		ls.Id = site.Id
		ls.Url = site.Url
		ls.Frequency = site.Frequency
		listProto = append(listProto, ls)
	}
	return &proto.ReadAllResponseSite{Sites: listProto}, nil
}

// Update site...
func (g *GRPCServer) Update(ctx context.Context, request *proto.UpdateRequestSite) (*proto.UpdateResponseSite, error) {
	g.log = logging.NewLoggers("server", "update")
	siteProto := request.GetSites()
	site := sites.Site{
		Id:        siteProto.GetId(),
		Url:       siteProto.GetUrl(),
		Frequency: siteProto.GetFrequency(),
		Deleted:   false,
	}
	if err := sites.UpdateSites(&site); err != nil {
		if err == sites.ErrSitesNotFound {
			err = status.Error(codes.NotFound, "unable to update")
			g.log.WarnLog().Str("when", "update site").Str("request", "failed to process").
				Err(err).Msg("unable to update  site")
		} else {
			err = status.Error(codes.Unknown, "unable to update")
			g.log.WarnLog().Str("when", "update site").Str("request", "failed to process").
				Err(err).Msg("unable to update site")
		}
		return nil, err
	}

	g.Backend.CreateOrUpdate(&site)
	return &proto.UpdateResponseSite{Updated: site.Id}, nil
}

// Delete site...
func (g *GRPCServer) Delete(ctx context.Context, request *proto.DeleteRequestSite) (*proto.DeleteResponseSite, error) {
	g.log = logging.NewLoggers("server", "delete")
	site := sites.Site{Id: request.GetId()}
	if err := sites.ReadSites(&site); err != nil {
		err = status.Error(codes.NotFound, "unable to delete")
		return nil, err
	}
	if err := sites.DeleteSites(&site); err != nil {
		if err == sites.ErrSitesNotFound {
			err = status.Error(codes.NotFound, "unable to delete")
			g.log.WarnLog().Str("when", "delete site").Str("request", "failed to process").
				Err(err).Msg("unable to delete site")
		} else {
			err = status.Error(codes.Unknown, "unable to delete")
			g.log.ErrorLog().Str("when", "delete site").Str("request", "failed to process").
				Err(err).Msg("unable to delete site")
		}
		return nil, err
	}

	g.Backend.Delete(&site)
	return &proto.DeleteResponseSite{Deleted: site.Id}, nil
}

func (g GRPCServer) ReadStatus(ctx context.Context, req *proto.ReadRequestState) (*proto.StatusResponse, error) {
	g.log = logging.NewLoggers("server", "readStatus")
	url := req.GetUrl()
	count := req.GetCount()
	list, err := statuses.ReadStatus(url, count)
	if err != nil {
		if err == statuses.ErrStatusNotFound {
			err = status.Error(codes.NotFound, "unable to get statuses")
			g.log.WarnLog().Str("when", "getting statuses").Str("request", "failed to process").
				Err(err).Msg("unable to get statuses")
		} else {
			err = status.Error(codes.Unknown, "unable to get statuses")
			g.log.ErrorLog().Str("when", "getting statuses").Str("request", "failed to process").
				Err(err).Msg("unable to get statuses")
		}
		return nil, err
	}

	return list, nil
}

// RunServer ...
func RunServer(cfg ServerConfig, ctx context.Context, server *GRPCServer, s *grpc.Server) error {
	server.log = logging.NewLoggers("server", "runServer")
	listen, err := net.Listen("tcp", cfg.GetServerAddress())
	if err != nil {
		err = status.Error(codes.Internal, "error listening server")
		server.log.ErrorLog().Str("when", "listening server").Err(err).
			Msg("error listening server, exiting")
	}
	proto.RegisterSitesServiceServer(s, server)

	return s.Serve(listen)
}