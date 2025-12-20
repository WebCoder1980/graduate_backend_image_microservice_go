package handler

import (
	"context"
	"graduate_backend_image_microservice_go/internal/constant"
	"graduate_backend_image_microservice_go/internal/service"
	"log"
	"net/http"
	"os"
	"strconv"
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

	fileId, err := h.service.Post(file, handler.Filename)
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
