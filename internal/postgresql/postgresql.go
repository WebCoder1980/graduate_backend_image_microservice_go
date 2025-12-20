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
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &PostgreSQL{db: db}, nil
}
