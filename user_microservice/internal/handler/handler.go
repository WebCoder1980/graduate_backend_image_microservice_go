package handler

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"user_microservice/internal/model"
	"user_microservice/internal/service"
)

const prefix = "/api/v1/user"

type Handler struct {
	ctx     context.Context
	service *service.Service
}

func NewHandler(ctx context.Context) (*Handler, error) {
	serv, err := service.NewService(ctx)
	if err != nil {
		return nil, err
	}

	return &Handler{
		ctx:     ctx,
		service: serv,
	}, err
}

func (h *Handler) UserRegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			log.Panic(err)
		}

		var body model.UserRegisterRequest
		err = json.Unmarshal(data, &body)
		if err != nil {
			log.Panic(err)
		}

		err = h.service.UserRegisterPost(&body)
		if err != nil {
			log.Panic(err)
		}

		w.WriteHeader(http.StatusCreated)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (h *Handler) Start() {
	http.HandleFunc(prefix+"/register", h.UserRegisterHandler)

	log.Panic(http.ListenAndServe(":"+os.Getenv("handler_port"), nil))
}
