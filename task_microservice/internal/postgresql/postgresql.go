package postgresql

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"graduate_backend_task_microservice/internal/model"
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
		CREATE TABLE IF NOT EXISTS task(
			id BIGSERIAL PRIMARY KEY
		);
		CREATE TABLE IF NOT EXISTS image_status(
		    id BIGINT PRIMARY KEY,
		    name TEXT NOT NULL UNIQUE
		);
		CREATE TABLE IF NOT EXISTS image(
		    id BIGSERIAL PRIMARY KEY,
		    task_id BIGINT REFERENCES task(id) NOT NULL,
		    position INT NOT NULL,
		    name TEXT NOT NULL,
		    format TEXT NOT NULL,
		    status_id BIGINT REFERENCES image_status(id) NOT NULL
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

func (p *PostgreSQL) ImageGetByTaskId(taskId int64) ([]model.ImageInfo, error) {
	var result []model.ImageInfo

	rows, err := p.db.Query(`	
		SELECT id, name, format, task_id, position, status_id
		FROM image
		WHERE task_id = $1
		ORDER BY id
	`, taskId)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var cur model.ImageInfo

		err = rows.Scan(&cur.Id, &cur.Filename, &cur.Format, &cur.TaskId, &cur.Position, &cur.StatusId)
		if err != nil {
			return nil, err
		}

		result = append(result, cur)
	}

	return result, nil
}

func (p *PostgreSQL) TaskCreate() (int64, error) {
	row := p.db.QueryRow("INSERT INTO task DEFAULT VALUES RETURNING id")
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
		SET status_id = $1
		WHERE task_id=$2 AND position=$3
	`, imageStatus.StatusId, imageStatus.TaskId, imageStatus.Position)
	if err != nil {
		return err
	}

	return nil
}
