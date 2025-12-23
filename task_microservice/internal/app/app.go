package app

import (
	"context"
	"graduate_backend_task_microservice/internal/handler"
	"graduate_backend_task_microservice/internal/kafkaconsumer"
	"log"
	"sync"
)

func Run() {
	ctx := context.Background()
	var wg sync.WaitGroup

	appHandler, err := handler.NewHandler(ctx)
	if err != nil {
		log.Panic(err)
	}
	wg.Go(appHandler.Start)

	kafka, err := kafkaconsumer.NewConsumer(ctx)
	if err != nil {
		log.Panic(err)
	}

	wg.Go(kafka.Start)

	wg.Wait()
}
