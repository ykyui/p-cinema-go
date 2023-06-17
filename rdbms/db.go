package rdbms

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var db *sql.DB

func InitDb() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
	conn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("UserName"), os.Getenv("Password"), os.Getenv("Addr"), os.Getenv("Port"), os.Getenv("Database"))
	//連接MySQL
	DB, err := sql.Open("mysql", conn)
	if err != nil {
		fmt.Println("connection to mysql failed:", err)
		panic(err)
	}

	maxLifetime, _ := strconv.Atoi(os.Getenv("MaxLifetime"))
	maxOpenConns, _ := strconv.Atoi(os.Getenv("MaxOpenConns"))
	maxIdleConns, _ := strconv.Atoi(os.Getenv("MaxIdleConns"))
	DB.SetConnMaxLifetime(time.Duration(maxLifetime) * time.Second)
	DB.SetMaxOpenConns(maxOpenConns)
	DB.SetMaxIdleConns(maxIdleConns)
	db = DB
}

func Close() {
	db.Close()
}

func nullString(s string) sql.NullString {
	if len(strings.TrimSpace(s)) == 0 {
		return sql.NullString{}
	} else {
		return sql.NullString{String: s, Valid: true}
	}
}

func nullInt(i int) sql.NullInt64 {
	if i == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(i), Valid: true}
}

func TransactionsStart() (*sql.Tx, error) {
	return db.Begin()
}
