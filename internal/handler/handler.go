package handler

import (
	"context"
	"graduate_backend_image_microservice_go/internal/constant"
	"graduate_backend_image_microservice_go/internal/service"
	"log"
	"net/http"
	"os"
)

const prefix = "/api/v1/image"

type Handler struct {
	service *service.Service
}

func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	r.ParseMultipartForm(constant.FileMaxSize)

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Ошибка получения файла: ", http.StatusBadRequest)
		return
	}
	defer file.Close()

	h.service.Post(file, handler.Filename)
}

func NewHandler(ctx context.Context) *Handler {
	return &Handler{service: service.NewService(ctx)}
}

func (h *Handler) Start() {
	http.HandleFunc(prefix+"/", h.Post)

	log.Panic(http.ListenAndServe(":"+os.Getenv("handler_port"), nil))
}
