package kafka

import (
	"context"
	"github.com/segmentio/kafka-go"
	"graduate_backend_image_microservice_go/internal/constant"
	"io"
	"log"
	"mime/multipart"
	"os"
)

type Producer struct {
	ctx         context.Context
	kafkaWriter *kafka.Writer
}

func NewProducer(ctx context.Context) *Producer {
	kafkaWriter := kafka.Writer{
		Addr:       kafka.TCP(os.Getenv("kafka_address")),
		Topic:      TopicName,
		BatchBytes: constant.FileMaxSize,
	}

	return &Producer{
		ctx:         ctx,
		kafkaWriter: &kafkaWriter,
	}
}

func (p *Producer) Write(file multipart.File, filename string) {
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Panic(err)
	}

	bytesResult := append([]byte(filename), EndFileName...)
	bytesResult = append(bytesResult, fileBytes...)

	err = p.kafkaWriter.WriteMessages(p.ctx, kafka.Message{
		Value: bytesResult,
	})
	if err != nil {
		log.Panic(err)
	}
}
