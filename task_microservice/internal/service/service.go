package service

import (
	"context"
	"errors"
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

func (s *Service) Post(files *multipart.Form) (int64, error) {
	if files == nil {
		return -1, errors.New("файл отсутствует")
	}

	for _, v2 := range files.File["file"] {
		val, err := v2.Open()
		if err != nil {
			return -1, err
		}

		fileBytes, err := io.ReadAll(val)
		if err != nil {
			return -1, err
		}

		filename := v2.Filename

		taskId, err := s.postgresql.TaskCreate(filename)
		if err != nil {
			return -1, err
		}

		fileFormat := strings.ToLower(filename[strings.LastIndex(filename, ".")+1:])
		minioFilename := strconv.FormatInt(taskId, 10) + "." + fileFormat

		s.minioClient.Upsert(fileBytes, minioFilename)

		s.kafkaProducer.Write(minioFilename)

		return taskId, nil
	}
	return -1, errors.New("файл отсутствует")
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
