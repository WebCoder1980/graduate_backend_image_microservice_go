package postgresql

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"graduate_backend_task_microservice/internal/model"
	"log"
	"os"
	"strconv"
	"time"
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
		CREATE TABLE IF NOT EXISTS task(
			id BIGSERIAL PRIMARY KEY,
			created_dt TIMESTAMP NOT NULL,
			width INT NULL,
			height INT NULL,
			format TEXT NULL,
			quality REAL NULL
		);
		CREATE TABLE IF NOT EXISTS image_status(
		    id BIGINT PRIMARY KEY,
		    name TEXT NOT NULL UNIQUE
		);
		CREATE TABLE IF NOT EXISTS image(
		    id BIGSERIAL PRIMARY KEY,
		    image_processor_image_id BIGINT NULL,
		    task_id BIGINT REFERENCES task(id) NOT NULL,
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

func (p *PostgreSQL) TaskGetById(id int64) (model.TaskInfo, error) {
	row := p.db.QueryRow("SELECT id, created_dt, width, height, format, quality FROM task WHERE id = $1", id)
	var taskInfo model.TaskInfo
	err := row.Scan(&taskInfo.Id, &taskInfo.CreatedDT, &taskInfo.Width, &taskInfo.Height, &taskInfo.Format, &taskInfo.Quality)
	if err != nil {
		return model.TaskInfo{}, err
	}

	return taskInfo, nil
}

func (p *PostgreSQL) ImageGetByTaskId(taskId int64) ([]model.ImageInfo, error) {
	var result []model.ImageInfo

	rows, err := p.db.Query(`	
		SELECT id, image_processor_image_id, name, format, task_id, position, status_id, end_dt
		FROM image
		WHERE task_id = $1
		ORDER BY id
	`, taskId)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var cur model.ImageInfo
		var endDT sql.NullTime

		err = rows.Scan(&cur.Id, &cur.ImageProcessorImageId, &cur.Filename, &cur.Format, &cur.TaskId, &cur.Position, &cur.StatusId, &endDT)
		if err != nil {
			return nil, err
		}

		if endDT.Valid {
			cur.EndDT = endDT.Time
		}

		result = append(result, cur)
	}

	return result, nil
}

func (p *PostgreSQL) TaskCreate(width *int, height *int, format *string, quality *float64) (int64, error) {
	row := p.db.QueryRow(`
		INSERT INTO task
		(created_dt, width, height, format, quality)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`,
		time.Now(),
		width,
		height,
		format,
		quality,
	)
	var resultId int64
	err := row.Scan(&resultId)
	if err != nil {
		return -1, err
	}
	return resultId, nil
}

func (p *PostgreSQL) ImageCreate(imageInfo model.ImageInfo) (int64, error) {
	row := p.db.QueryRow("INSERT INTO image (task_id, position, name, format, status_id) VALUES ($1, $2, $3, $4, $5) RETURNING id", imageInfo.TaskId, imageInfo.Position, imageInfo.Filename, imageInfo.Format, imageInfo.StatusId)
	var resultId int64
	err := row.Scan(&resultId)
	if err != nil {
		return -1, err
	}
	return resultId, nil
}

func (p *PostgreSQL) ImageUpdateStatus(imageStatus model.ImageStatus) error {
	_, err := p.db.Exec(`
		UPDATE image
		SET
		    image_processor_image_id = $1,
		    status_id = $2,
			end_dt = $3
		WHERE task_id=$4 AND position=$5
	`, imageStatus.ImageProcessorImageId, imageStatus.StatusId, imageStatus.EndDT, imageStatus.TaskId, imageStatus.Position)
	if err != nil {
		return err
	}

	return nil
}
