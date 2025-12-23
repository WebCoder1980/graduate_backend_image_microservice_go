package app

import (
	"context"
	"graduate_backend_image_processor_microservice/internal/kafkaconsumer"
	"log"
)

func Run() {
	ctx := context.Background()

	consumer, err := kafkaconsumer.NewConsumer(ctx)
	if err != nil {
		log.Panic(err)
	}

	consumer.Start()
}
