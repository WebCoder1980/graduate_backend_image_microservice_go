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
		CREATE TABLE IF NOT EXISTS image_type(
		    id BIGSERIAL PRIMARY KEY,
		    name TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS image(
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			type_id BIGINT REFERENCES image_type(id) NOT NULL
		);
	`)
	if err != nil {
		return err
	}

	row := p.db.QueryRow("SELECT COUNT(*) FROM image_type")
	var count int64
	err = row.Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		_, err = p.db.Exec("INSERT INTO image_type(name) VALUES ('Обрабатывается'), ('Успех'), ('Ошибка')")
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PostgreSQL) CreateImage(name string) (int64, error) {
	queryRes := p.db.QueryRow("SELECT id FROM image_type WHERE name = 'Обрабатывается'")
	var id int64
	err := queryRes.Scan(&id)
	if err != nil {
		return -1, err
	}

	queryRes = p.db.QueryRow("INSERT INTO image(name, type_id) VALUES($1, $2) RETURNING id", name, id)
	var resultId int64
	err = queryRes.Scan(&resultId)
	if err != nil {
		return -1, err
	}

	return resultId, nil
}
