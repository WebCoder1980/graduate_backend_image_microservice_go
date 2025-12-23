package postgresql

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
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
		CREATE TABLE IF NOT EXISTS task_status(
		    id BIGSERIAL PRIMARY KEY,
		    name TEXT NOT NULL UNIQUE
		);
		CREATE TABLE IF NOT EXISTS task(
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			status_id BIGINT REFERENCES task_status(id) NOT NULL
		);
	`)
	if err != nil {
		return err
	}

	queryRowRes := p.db.QueryRow("SELECT COUNT(*) FROM task_status")
	var count int64
	err = queryRowRes.Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		_, err = p.db.Exec("INSERT INTO task_status(name) VALUES ('Обрабатывается'), ('Успех'), ('Ошибка')")
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PostgreSQL) TaskCreate(name string) (int64, error) {
	taskTypeId, err := p.TaskStatusByName("Обрабатывается")
	if err != nil {
		return -1, err
	}

	row := p.db.QueryRow("INSERT INTO task (name, status_id) VALUES ($1, $2) RETURNING id", name, taskTypeId)
	var resultId int64
	err = row.Scan(&resultId)
	if err != nil {
		return -1, err
	}
	return resultId, nil
}

func (p *PostgreSQL) TaskUpdateStatus(taskId int64, statusId int64) error {
	_, err := p.db.Exec(`
		UPDATE task
		SET status_id = $1
		WHERE id = $2
	`, statusId, taskId)
	if err != nil {
		return err
	}

	return nil
}

func (p *PostgreSQL) TaskStatusByName(name string) (int64, error) {
	row := p.db.QueryRow("SELECT id FROM task_status WHERE name=$1", name)
	var result int64
	err := row.Scan(&result)
	if err != nil {
		return -1, err
	}
	return result, nil
}
