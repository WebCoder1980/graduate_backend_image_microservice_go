package handler

import (
	"context"
	"graduate_backend_task_microservice/internal/constant"
	"graduate_backend_task_microservice/internal/service"
	"log"
	"net/http"
	"os"
	"strconv"
)

const prefix = "/api/v1/task"

type Handler struct {
	service *service.Service
}

func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	r.ParseMultipartForm(constant.FileMaxSize)

	files := r.MultipartForm

	fileId, err := h.service.Post(files)
	if err != nil {
		log.Panic(err)
	}
	_, err = w.Write([]byte(strconv.FormatInt(fileId, 10)))
	if err != nil {
		log.Panic(err)
	}
}

func NewHandler(ctx context.Context) (*Handler, error) {
	service, err := service.NewService(ctx)
	if err != nil {
		return nil, err
	}

	return &Handler{service: service}, nil
}

func (h *Handler) Start() {
	http.HandleFunc(prefix+"/", h.Post)

	log.Panic(http.ListenAndServe(":"+os.Getenv("handler_port"), nil))
}
