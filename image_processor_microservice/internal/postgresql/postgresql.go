package postgresql

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"graduate_backend_image_processor_microservice/internal/model"
	"log"
	"os"
	"strconv"
)

type PostgreSQL struct {
	db *sql.DB
}

func NewPostgreSQL() (*PostgreSQL, error) {
	port, err := strconv.Atoi(os.Getenv("postgresql_port"))
	if err != nil {
		log.Panic(err)
	}

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("postgresql_host"),
		port,
		os.Getenv("postgresql_user"),
		os.Getenv("postgresql_password"),
		os.Getenv("postgresql_dbname"),
		"disable",
	)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, err
	}

	result := &PostgreSQL{db: db}
	err = result.init()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (p *PostgreSQL) init() error {
	if err := p.db.Ping(); err != nil {
		return err
	}

	_, err := p.db.Exec(`
		CREATE TABLE IF NOT EXISTS image_status(
		    id BIGINT PRIMARY KEY,
		    name TEXT NOT NULL UNIQUE
		);
		CREATE TABLE IF NOT EXISTS image(
		    id BIGSERIAL PRIMARY KEY,
		    task_id BIGINT NOT NULL,
		    position INT NOT NULL,
		    name TEXT NOT NULL,
		    format TEXT NOT NULL,
		    status_id BIGINT REFERENCES image_status(id) NOT NULL,
		    end_dt TIMESTAMP NULL,
		    CONSTRAINT uq_image_task_id_position UNIQUE (task_id, position)
		)
	`)
	if err != nil {
		return err
	}

	queryRowRes := p.db.QueryRow("SELECT COUNT(*) FROM image_status")
	var count int64
	err = queryRowRes.Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		_, err = p.db.Exec("INSERT INTO image_status(id, name) VALUES (1, 'Обрабатывается'), (2, 'Успех'), (3, 'Ошибка')")
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PostgreSQL) ImageGetByid(imageId int64) (model.ImageInfo, error) {
	row := p.db.QueryRow(`	
		SELECT id, name, format, task_id, position, status_id, end_dt
		FROM image
		WHERE id = $1
	`, imageId)

	var result model.ImageInfo

	err := row.Scan(&result.Id, &result.Filename, &result.Format, &result.TaskId, &result.Position, &result.StatusId, &result.EndDT)
	if err != nil {
		return model.ImageInfo{}, err
	}

	return result, nil
}

func (p *PostgreSQL) ImageCreate(imageInfo model.ImageInfo) (int64, error) {
	row := p.db.QueryRow("INSERT INTO image (task_id, position, name, format, status_id, end_dt) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id", imageInfo.TaskId, imageInfo.Position, imageInfo.Filename, imageInfo.Format, imageInfo.StatusId, imageInfo.EndDT)
	var resultId int64
	err := row.Scan(&resultId)
	if err != nil {
		return -1, err
	}
	return resultId, nil
}
