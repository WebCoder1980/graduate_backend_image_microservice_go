package service

import (
	"context"
	"graduate_backend_image_microservice_go/internal/kafka"
	"graduate_backend_image_microservice_go/internal/postgresql"
	"mime/multipart"
	"strconv"
	"strings"
)

type Service struct {
	ctx           context.Context
	kafkaProducer *kafka.Producer
	postgresql    *postgresql.PostgreSQL
}

func NewService(ctx context.Context) (*Service, error) {
	psql, err := postgresql.NewPostgreSQL()
	if err != nil {
		return nil, err
	}
	return &Service{
		ctx:           ctx,
		kafkaProducer: kafka.NewProducer(ctx),
		postgresql:    psql,
	}, nil
}

func (s *Service) Post(file multipart.File, filename string) (int64, error) {
	imageId, err := s.postgresql.CreateImage(filename)
	if err != nil {
		return -1, err
	}

	fileFormat := strings.ToLower(filename[strings.LastIndex(filename, ".")+1:])
	minioFilename := strconv.FormatInt(imageId, 10) + "." + fileFormat

	s.kafkaProducer.Write(file, minioFilename)

	return imageId, nil
}
