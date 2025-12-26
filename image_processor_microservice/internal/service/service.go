package service

import (
	"context"
	"errors"
	"fmt"
	"graduate_backend_image_processor_microservice/internal/constant"
	"graduate_backend_image_processor_microservice/internal/kafkaproducer"
	"graduate_backend_image_processor_microservice/internal/minio"
	"graduate_backend_image_processor_microservice/internal/model"
	"graduate_backend_image_processor_microservice/internal/postgresql"
	"strconv"
	"time"
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

func (s *Service) ImageGetById(imageId int64) ([]byte, error) {
	imageInfo, err := s.postgresql.ImageGetByid(imageId)
	if err != nil {
		return nil, err
	}

	minioFilename := fmt.Sprintf("%d_%d.%s", imageInfo.TaskId, imageInfo.Position, imageInfo.Format)

	data, err := s.minioClient.Get(minio.BucketTargetName, minioFilename)
	if err != nil {
		return nil, err
	}

	return data, err
}

func (s *Service) ImageProcessor(imageInfo *model.ImageInfo) error {
	imageInfo.StatusId = constant.StatusSuccessful

	minioFilename := strconv.FormatInt(imageInfo.TaskId, 10) + "_" + strconv.Itoa(imageInfo.Position) + "." + imageInfo.Format

	source, err := s.minioClient.Get(minio.BucketSourceName, minioFilename)
	if err != nil {
		imageInfo.StatusId = constant.StatusFailed
		imageInfo.EndDT = time.Now()
		err2 := s.ImageProcessorKafkaWrite(imageInfo)
		if err2 != nil {
			return errors.New(err.Error() + "; " + err2.Error())
		}
		return nil
	}

	// TODO processing

	imageId, err := s.postgresql.ImageCreate(*imageInfo)
	if err != nil {
		imageInfo.StatusId = constant.StatusFailed
		imageInfo.EndDT = time.Now()
		err2 := s.ImageProcessorKafkaWrite(imageInfo)
		if err2 != nil {
			return errors.New(err.Error() + "; " + err2.Error())
		}
		return nil
	}
	imageInfo.Id = imageId

	err = s.minioClient.Upsert(source, minioFilename)
	if err != nil {
		imageInfo.StatusId = constant.StatusFailed
		imageInfo.EndDT = time.Now()
		err2 := s.ImageProcessorKafkaWrite(imageInfo)
		if err2 != nil {
			return errors.New(err.Error() + "; " + err2.Error())
		}
		return nil
	}

	imageInfo.EndDT = time.Now()
	err = s.ImageProcessorKafkaWrite(imageInfo)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ImageProcessorKafkaWrite(imageInfo *model.ImageInfo) error {
	imageStatus := model.ImageStatus{
		TaskId:   imageInfo.TaskId,
		Position: imageInfo.Position,
		StatusId: imageInfo.StatusId,
		EndDT:    imageInfo.EndDT,
	}

	err := s.kafkaProducer.Write(imageStatus)
	if err != nil {
		return err
	}

	return nil
}
