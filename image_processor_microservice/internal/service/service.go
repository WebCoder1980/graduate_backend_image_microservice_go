package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/disintegration/imaging"
	"graduate_backend_image_processor_microservice/internal/constant"
	"graduate_backend_image_processor_microservice/internal/kafkaproducer"
	"graduate_backend_image_processor_microservice/internal/minio"
	"graduate_backend_image_processor_microservice/internal/model"
	"graduate_backend_image_processor_microservice/internal/postgresql"
	"image/jpeg"
	"image/png"
	"strconv"
	"time"
)

const (
	FormatJPEG string = "jpg"
	FormatPNG  string = "png"
	FormatWEBP string = "webp"
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

func (s *Service) ServiceImageProcessor(imageRequest *model.ImageRequest) error {
	imageRequest.StatusId = constant.StatusSuccessful

	minioFilename := strconv.FormatInt(imageRequest.TaskId, 10) + "_" + strconv.Itoa(imageRequest.Position) + "." + imageRequest.Format

	sourceBytes, err := s.minioClient.Get(minio.BucketSourceName, minioFilename)
	if err != nil {
		imageRequest.StatusId = constant.StatusFailed
		imageRequest.EndDT = time.Now()
		err2 := s.ImageProcessorKafkaWrite(imageRequest)
		if err2 != nil {
			return errors.New(err.Error() + "; " + err2.Error())
		}
		return nil
	}

	targetBytes, err := s.ImageProcess(sourceBytes, imageRequest.TargetFormat, imageRequest.Width, imageRequest.Height, imageRequest.Quality)
	if err != nil {
		return err
	}
	imageRequest.EndDT = time.Now()

	imageId, err := s.postgresql.ImageCreate(*imageRequest)
	if err != nil {
		imageRequest.StatusId = constant.StatusFailed
		err2 := s.ImageProcessorKafkaWrite(imageRequest)
		if err2 != nil {
			return errors.New(err.Error() + "; " + err2.Error())
		}
		return nil
	}
	imageRequest.Id = imageId

	err = s.minioClient.Upsert(targetBytes, minioFilename)
	if err != nil {
		imageRequest.StatusId = constant.StatusFailed
		err2 := s.ImageProcessorKafkaWrite(imageRequest)
		if err2 != nil {
			return errors.New(err.Error() + "; " + err2.Error())
		}
		return nil
	}

	err = s.ImageProcessorKafkaWrite(imageRequest)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ImageProcessorKafkaWrite(imageRequest *model.ImageRequest) error {
	imageStatus := model.ImageStatus{
		TaskId:   imageRequest.TaskId,
		Position: imageRequest.Position,
		StatusId: imageRequest.StatusId,
		EndDT:    imageRequest.EndDT,
	}

	err := s.kafkaProducer.Write(imageStatus)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ImageProcess(input []byte, format *string, width *int, height *int, quality *float32) ([]byte, error) {
	img, err := imaging.Decode(bytes.NewReader(input))
	if err != nil {
		return nil, err
	}

	if width != nil && height != nil {
		img = imaging.Resize(img, *width, *height, imaging.Lanczos)
	}

	var buf bytes.Buffer

	switch *format {
	case FormatJPEG:
		if quality == nil {
			var q float32 = 0.9
			quality = &q
		}
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: int(*quality * 100)})

	case FormatPNG:
		err = png.Encode(&buf, img)

	case FormatWEBP:
		if quality == nil {
			var q float32 = 0.9
			quality = &q
		}
		//err = webp.Encode(&buf, img, &webp.Options{
		//	Lossless: false,
		//	Quality:  quality,
		//})

	default:
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
