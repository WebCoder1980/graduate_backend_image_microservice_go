package main

import (
	"bytes"
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/segmentio/kafka-go"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

const prefix = "/api/v1/image"

func worker() {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "image-topic",
		GroupID: "group0",
	})
	defer reader.Close()

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Fatal("Ошибка при получении:", err)
		}

		msgStr := string(msg.Value)

		offset := strings.Index(msgStr, "---END FILE NAME---")

		ctx := context.Background()
		endpoint := "localhost:9000"
		accessKeyID := "minioadmin"
		secretAccessKey := "minioadmin"
		useSSL := false

		minioClient, err := minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
			Secure: useSSL,
		})
		if err != nil {
			log.Fatal(err)
		}

		bucketName := "filebuckit"

		exists, err := minioClient.BucketExists(ctx, bucketName)
		if err != nil {
			log.Fatal(err)
		}
		if !exists {
			err := minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
			if err != nil {
				log.Fatal(err)
			}
		}

		reader := bytes.NewReader(msg.Value[offset+(len("---END FILE NAME---")):])
		_, err = minioClient.PutObject(ctx, bucketName, msgStr[:offset], reader, int64(len(msg.Value[offset+(len("---END FILE NAME---")):])), minio.PutObjectOptions{})
		if err != nil {
			log.Fatal(err)
		}
	}
}

func handler() {
	http.HandleFunc(prefix+"/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		r.ParseMultipartForm(10 << 20)

		file, handler, err := r.FormFile("file")

		if err != nil {
			http.Error(w, "Ошибка получения файла", http.StatusBadRequest)
			return
		}
		defer file.Close()

		ctx := context.Background()

		writer := kafka.NewWriter(kafka.WriterConfig{
			Brokers: []string{"localhost:9092"},
			Topic:   "image-topic",
		})
		defer writer.Close()

		fileBytes, err := io.ReadAll(file)

		bytesResult := append([]byte(handler.Filename), "---END FILE NAME---"...)
		bytesResult = append(bytesResult, fileBytes...)

		err = writer.WriteMessages(ctx, kafka.Message{
			Value: bytesResult,
		})
		if err != nil {
			log.Fatal("Ошибка при отправке:", err)
		}
	})

	log.Fatal(http.ListenAndServe(":5267", nil))
}

func main() {
	var wg sync.WaitGroup

	wg.Go(handler)
	wg.Go(worker)

	wg.Wait()
}
