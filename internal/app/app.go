package app

import (
	"context"
	"graduate_backend_image_microservice_go/internal/handler"
	"graduate_backend_image_microservice_go/internal/kafka"
	"log"
	"sync"
)

func Run() {
	ctx := context.Background()
	var wg sync.WaitGroup

	kafkaConsumer, err := kafka.NewConsumer(ctx)
	if err != nil {
		log.Panic(err)
	}
	wg.Go(kafkaConsumer.Start)

	appHandler := handler.NewHandler(ctx)
	wg.Go(appHandler.Start)

	wg.Wait()
}
