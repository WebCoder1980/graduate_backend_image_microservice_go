package kafka

import (
	"context"
	"github.com/segmentio/kafka-go"
	"graduate_backend_image_microservice_go/internal/minio"
	"log"
	"os"
	"strings"
)

type Consumer struct {
	ctx         context.Context
	kafkaReader *kafka.Reader
	minioClient *minio.Client
}

const consumerGroup = "group0"

func NewConsumer(ctx context.Context) (*Consumer, error) {
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{os.Getenv("kafka_address")},
		Topic:   TopicName,
		GroupID: consumerGroup,
	})

	minioClient, err := minio.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		ctx:         ctx,
		kafkaReader: kafkaReader,
		minioClient: minioClient,
	}, nil
}

func (c *Consumer) Start() {
	for {
		msg, err := c.kafkaReader.ReadMessage(c.ctx)
		if err != nil {
			log.Panic(err)
		}

		msgStr := string(msg.Value)
		offset := strings.Index(msgStr, EndFileName)
		file := msg.Value[(offset + len(EndFileName)):]
		filename := msgStr[:offset]

		c.minioClient.Upsert(filename, file)
	}
}
