package server

import (
	"github.com/UndeadDemidov/yandex-praktikum/internal/app/handlers"
	pb "github.com/UndeadDemidov/yandex-praktikum/proto"
	"google.golang.org/grpc"
)

func NewGRPCServer(baseURL string, repo handlers.Repository) *grpc.Server {
	linkStore := repo
	handler := handlers.NewURLShortener(baseURL, linkStore)

	s := grpc.NewServer()
	pb.RegisterShortenerServer(s, handler)
	return s
}
