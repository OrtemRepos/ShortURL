package grpc

import (
	"context"
	"errors"

	"github.com/OrtemRepos/ShortURL/shortener-service/gen/url"
	"github.com/OrtemRepos/ShortURL/shortener-service/internal/controller"
	"github.com/OrtemRepos/ShortURL/shortener-service/internal/domain"
	"github.com/OrtemRepos/ShortURL/shortener-service/internal/repository"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Handler struct {
	url.UnimplementedShortenerServiceServer
	ctrl   *controller.Controller
	logger *zap.Logger
}

func New(ctrl *controller.Controller, logger *zap.Logger) *Handler {
	return &Handler{
		ctrl:   ctrl,
		logger: logger,
	}
}

func (h *Handler) GetOriginalURL(ctx context.Context, req *url.ShortURL) (*url.OriginalURL, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "Nil req")
	}
	h.logger.Info("got request", zap.Any("req", req))

	urls, err := h.ctrl.Get(ctx, req.Url)
	if errors.Is(err, repository.ErrURLNotFound) {
		return nil, status.Error(codes.NotFound, "URL not found")
	} else if err != nil {
		h.logger.Error("failed to get url", zap.Error(err), zap.String("short_url", req.Url))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &url.OriginalURL{Url: urls.OriginalURL}, nil
}

func (h *Handler) GenerateShortURL(ctx context.Context, req *url.GenerateShortURLRequest) (*url.URL, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "Nil req")
	}
	h.logger.Info("got request", zap.Any("req", req))
	domainURL := domain.NewURL(req.OriginalUrl)
	err := h.ctrl.Save(ctx, domainURL, req.Ttl.AsDuration())
	if err != nil {
		h.logger.Error("failed to get url", zap.Error(err), zap.String("original_url", req.OriginalUrl))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &url.URL{OriginalUrl: domainURL.OriginalURL, ShortUrl: domainURL.ShortURL}, nil
}

func (h *Handler) DeleteShortURL(ctx context.Context, req *url.ShortURL) (*emptypb.Empty, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "Nil req")
	}
	h.logger.Info("got request", zap.Any("req", req))

	err := h.ctrl.Delete(ctx, req.Url)
	if err != nil {
		h.logger.Error("failed to delete url", zap.Error(err), zap.String("short_url", req.Url))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}
