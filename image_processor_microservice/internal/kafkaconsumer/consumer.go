package kafkaconsumer

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"graduate_backend_image_processor_microservice/internal/model"
	"graduate_backend_image_processor_microservice/internal/service"
	"log"
	"os"
)

const (
	TopicTaskRequest = "task_request"
)

type Consumer struct {
	ctx         context.Context
	kafkaReader *kafka.Reader
	service     *service.Service
}

const consumerGroup = "group0"

func NewConsumer(ctx context.Context) (*Consumer, error) {
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{os.Getenv("kafka_address")},
		Topic:   TopicTaskRequest,
		GroupID: consumerGroup,
	})

	serv, err := service.NewService(ctx)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		ctx:         ctx,
		kafkaReader: kafkaReader,
		service:     serv,
	}, nil
}

func (c *Consumer) Start() {
	for {
		msg, err := c.kafkaReader.ReadMessage(c.ctx)
		if err != nil {
			log.Panic(err)
		}

		var imageRequest model.ImageRequest
		err = json.Unmarshal(msg.Value, &imageRequest)
		if err != nil {
			log.Panic(err)
		}

		err = c.service.ServiceImageProcessor(&imageRequest)
		if err != nil {
			log.Panic(err)
		}
	}
}
