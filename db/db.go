package db

import (
	"NourishNestApp/logger"
	"database/sql"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/libsql/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

var Db *sql.DB

func Init() {
	dbUrl := os.Getenv("DB_URL")
	initDb(dbUrl)
	migrateDb(Db)
}

func initDb(filepath string) *sql.DB {
	conn, err := sql.Open("libsql", filepath)
	if err != nil {
		panic(err)
	}

	// Test the connection.
	if err := conn.Ping(); err != nil {
		conn.Close()
		panic(err)
	}

	logger.Log.Info(fmt.Sprintf("Connected to db: '%s'", filepath))
	Db = conn

	return nil
}

func migrateDb(db *sql.DB) {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		panic(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"sqlite3", driver)
	if err != nil {
		panic(err)
	}

	logger.Log.Info("Running DB migrations")
	m.Up()
	version, _, _ := m.Version()

	logger.Log.Info("DB migrations completed")
	logger.Log.Info(fmt.Sprintf("Current migration version: '%d'", version))
	logger.Log.Info("Initialised DB")
}
