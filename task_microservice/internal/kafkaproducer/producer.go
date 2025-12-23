package kafkaproducer

import (
	"context"
	"github.com/segmentio/kafka-go"
	"log"
	"os"
)

const (
	TopicName = "task_request"
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
		Topic: TopicName,
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
		Topic:             TopicName,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
	if err != nil {
		return err
	}

	return nil
}

func (p *Producer) Write(filename string) {
	err := p.kafkaWriter.WriteMessages(p.ctx, kafka.Message{
		Value: []byte(filename),
	})
	if err != nil {
		log.Panic(err)
	}
}
