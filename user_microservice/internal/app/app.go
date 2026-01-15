package app

import (
	"context"
	"log"
	"user_microservice/internal/handler"
)

func Run() {
	ctx := context.Background()

	hand, err := handler.NewHandler(ctx)
	if err != nil {
		log.Panic(err)
	}
	hand.Start()
}
