package kafkaproducer

import (
	"context"
	"github.com/segmentio/kafka-go"
	"log"
	"os"
	"strconv"
)

const (
	TopicImage = "image_upsert"
)

type Producer struct {
	ctx         context.Context
	kafkaWriter *kafka.Writer
}

func NewProducer(ctx context.Context) (*Producer, error) {
	err := initTopic()
	if err != nil {
		return nil, err
	}

	kafkaWriter := kafka.Writer{
		Addr:  kafka.TCP(os.Getenv("kafka_address")),
		Topic: TopicImage,
	}

	return &Producer{
		ctx:         ctx,
		kafkaWriter: &kafkaWriter,
	}, nil
}

func initTopic() error {
	conn, err := kafka.Dial("tcp", os.Getenv("kafka_address"))
	if err != nil {
		return err
	}
	defer conn.Close()

	err = conn.CreateTopics(kafka.TopicConfig{
		Topic:             TopicImage,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
	if err != nil {
		return err
	}

	return nil
}

func (p *Producer) Write(imageId int64) {
	err := p.kafkaWriter.WriteMessages(p.ctx, kafka.Message{
		Value: []byte(strconv.FormatInt(imageId, 10)),
	})
	if err != nil {
		log.Panic(err)
	}
}
