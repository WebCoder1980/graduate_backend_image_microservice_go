package app

import (
	"context"
	"graduate_backend_image_processor_microservice/internal/handler"
	"graduate_backend_image_processor_microservice/internal/kafkaconsumer"
	"log"
	"sync"
)

func Run() {
	ctx := context.Background()

	wg := sync.WaitGroup{}

	appHandler, err := handler.NewHandler(ctx)
	if err != nil {
		log.Panic(err)
	}
	wg.Go(appHandler.Start)

	consumer, err := kafkaconsumer.NewConsumer(ctx)
	if err != nil {
		log.Panic(err)
	}
	wg.Go(consumer.Start)

	wg.Wait()
}
