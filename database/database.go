package database

import (
	"github.com/jackc/pgx"
)

var DB *pgx.ConnPool

const maxConn = 50

func Connect() (conn *pgx.ConnPool) {
	connConfig := pgx.ConnConfig{
		User:     "ksu",
		Password: "pswd",
		Host:     "localhost",
		Port:     5432,
		Database: "parkdb",
	}

	DB, _ = pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:     connConfig,
		MaxConnections: maxConn,
	})

	return DB
}
