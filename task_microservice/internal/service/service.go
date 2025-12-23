package service

import (
	"context"
	"graduate_backend_task_microservice/internal/kafkaproducer"
	"graduate_backend_task_microservice/internal/minio"
	"graduate_backend_task_microservice/internal/postgresql"
	"io"
	"mime/multipart"
	"strconv"
	"strings"
)

type Service struct {
	ctx           context.Context
	kafkaProducer *kafkaproducer.Producer
	minioClient   *minio.Client
	postgresql    *postgresql.PostgreSQL
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
		kafkaProducer: kafka,
		minioClient:   minioClient,
		postgresql:    psql,
	}, nil
}

func (s *Service) Post(file multipart.File, filename string) (int64, error) {
	imageId, err := s.postgresql.TaskCreate(filename)
	if err != nil {
		return -1, err
	}

	fileFormat := strings.ToLower(filename[strings.LastIndex(filename, ".")+1:])
	minioFilename := strconv.FormatInt(imageId, 10) + "." + fileFormat

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return -1, err
	}
	s.minioClient.Upsert(fileBytes, minioFilename)

	s.kafkaProducer.Write(minioFilename)

	return imageId, nil
}

func (s *Service) TaskUpdateStatus(taskId int64) error {
	statusId, err := s.postgresql.TaskStatusByName("Успех")
	if err != nil {
		return err
	}

	err = s.postgresql.TaskUpdateStatus(taskId, statusId)
	if err != nil {
		return err
	}

	return nil
}
