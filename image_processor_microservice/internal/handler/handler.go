package handler

import (
	"context"
	"graduate_backend_image_processor_microservice/internal/service"
	"log"
	"net/http"
	"os"
	"strconv"
)

const prefix = "/api/v1/image-processor"

type Handler struct {
	service *service.Service
}

func NewHandler(ctx context.Context) (*Handler, error) {
	serv, err := service.NewService(ctx)
	if err != nil {
		return nil, err
	}

	return &Handler{service: serv}, nil
}

func (h *Handler) ImageIdGet(w http.ResponseWriter, r *http.Request) {
	imageId, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		log.Panic(err)
	}

	data, err := h.service.ImageGetById(imageId)
	if err != nil {
		log.Panic(err)
	}

	_, err = w.Write(data)
	if err != nil {
		log.Panic(err)
	}
}

func (h *Handler) ImageIdHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ImageIdGet(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) Start() {
	http.HandleFunc(prefix+"/{id}", h.ImageIdHandler)

	log.Panic(http.ListenAndServe(":"+os.Getenv("handler_port"), nil))
}
