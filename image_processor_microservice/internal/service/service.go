package service

import (
	"context"
	"graduate_backend_image_processor_microservice/internal/kafkaproducer"
	"graduate_backend_image_processor_microservice/internal/minio"
	"graduate_backend_image_processor_microservice/internal/postgresql"
	"strconv"
	"strings"
)

type Service struct {
	ctx           context.Context
	postgresql    *postgresql.PostgreSQL
	minioClient   *minio.Client
	kafkaProducer *kafkaproducer.Producer
}

func NewService(ctx context.Context) (*Service, error) {
	psql, err := postgresql.NewPostgreSQL()
	if err != nil {
		return nil, err
	}

	minioClient, err := minio.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	kafka, err := kafkaproducer.NewProducer(ctx)
	if err != nil {
		return nil, err
	}

	return &Service{
		ctx:           ctx,
		postgresql:    psql,
		minioClient:   minioClient,
		kafkaProducer: kafka,
	}, nil
}

func (s *Service) ImageProcessor(filename string) error {
	source, err := s.minioClient.Get(filename)
	if err != nil {
		return err
	}

	err = s.minioClient.Upsert(source, filename)
	if err != nil {
		return err
	}

	seperator := strings.LastIndex(filename, ".")

	imageId, err := strconv.ParseInt(filename[:seperator], 10, 64)

	s.kafkaProducer.Write(imageId)

	return nil
}
