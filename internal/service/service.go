package service

import (
	"context"
	"graduate_backend_image_microservice_go/internal/kafka"
	"mime/multipart"
)

type Service struct {
	ctx           context.Context
	kafkaProducer *kafka.Producer
}

func NewService(ctx context.Context) *Service {
	return &Service{
		ctx:           ctx,
		kafkaProducer: kafka.NewProducer(ctx),
	}
}

func (s *Service) Post(file multipart.File, filename string) {
	s.kafkaProducer.Write(file, filename)
}
