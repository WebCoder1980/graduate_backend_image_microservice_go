package kafkaconsumer

import (
	"context"
	"github.com/segmentio/kafka-go"
	"graduate_backend_task_microservice/internal/service"
	"log"
	"os"
	"strconv"
)

const (
	TopicTaskRequest = "image_upsert"
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

		taskId, err := strconv.ParseInt(string(msg.Value), 10, 64)
		if err != nil {
			log.Panic(err)
		}
		err = c.service.TaskUpdateStatus(taskId)
		if err != nil {
			log.Panic(err)
		}
	}
}
