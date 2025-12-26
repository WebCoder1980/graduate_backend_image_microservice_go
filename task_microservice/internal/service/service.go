package service

import (
	"context"
	"errors"
	"graduate_backend_task_microservice/internal/constant"
	"graduate_backend_task_microservice/internal/kafkaproducer"
	"graduate_backend_task_microservice/internal/minio"
	"graduate_backend_task_microservice/internal/model"
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

func (s *Service) GetImagesByTaskId(taskId int64) (model.TaskResponse, error) {
	images, err := s.postgresql.ImageGetByTaskId(taskId)
	if err != nil {
		return model.TaskResponse{}, err
	}

	var commonStatusId int64 = constant.StatusSuccessful

	isWork, isFailed := false, false

	for _, val := range images {
		switch val.StatusId {
		case constant.StatusInWork:
			isWork = true
		case constant.StatusFailed:
			isFailed = true
			break
		default:
		}
	}

	if isWork {
		commonStatusId = constant.StatusInWork
	}
	if isFailed {
		commonStatusId = constant.StatusFailed
	}

	taskInfo, err := s.postgresql.TaskGetById(taskId)
	if err != nil {
		return model.TaskResponse{}, err
	}

	return model.TaskResponse{
		CommonStatusId: commonStatusId,
		Images:         images,
		CreatedDT:      taskInfo.CreatedDT,
	}, nil
}

func (s *Service) Post(files *multipart.Form) (int64, error) {
	if files == nil {
		return -1, errors.New("файл отсутствует")
	}

	taskId, err := s.postgresql.TaskCreate()
	if err != nil {
		return -1, nil
	}

	for i, v2 := range files.File["file"] {
		imageInfo := model.ImageInfo{
			TaskId:   taskId,
			Position: i + 1,
			StatusId: constant.StatusInWork,
		}

		formatSeparator := strings.LastIndex(v2.Filename, ".")
		imageInfo.Filename = v2.Filename[:formatSeparator]
		imageInfo.Format = strings.ToLower(v2.Filename[formatSeparator+1:])

		imageId, err := s.postgresql.ImageCreate(imageInfo)
		if err != nil {
			return -1, err
		}
		imageInfo.Id = imageId

		val, err := v2.Open()
		if err != nil {
			return -1, err
		}

		fileBytes, err := io.ReadAll(val)
		if err != nil {
			return -1, err
		}

		minioFilename := strconv.FormatInt(imageInfo.TaskId, 10) + "_" + strconv.Itoa(imageInfo.Position) + "." + imageInfo.Format

		s.minioClient.Upsert(fileBytes, minioFilename)

		err = s.kafkaProducer.Write(&imageInfo)
		if err != nil {
			return -1, err
		}
	}
	return taskId, nil
}

func (s *Service) TaskUpdateStatus(imageStatus model.ImageStatus) error {
	err := s.postgresql.ImageUpdateStatus(imageStatus)

	if err != nil {
		return err
	}

	return nil
}
