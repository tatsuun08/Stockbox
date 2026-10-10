package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"net/http"

	_ "modernc.org/sqlite"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	// "stockbox/db"
)

//go:embed sql/schema.sql
var ddl string

// DB接続
func connect_db() (*sql.DB, error) {
	// データベース接続
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil{
		return nil, err
	}
	return db, nil
}

// テーブル作成
func create_table(db *sql.DB, ctx context.Context) error {
	if _, err := db.ExecContext(ctx, ddl); err != nil {
		return err
	}
	return nil
}

func run() error {
	var port int = 1323
	
	ctx := context.Background()

	db, err := connect_db()
	if err != nil{
		return err
	}

	err = create_table(db, ctx)
	if err != nil{
		return err
	}

	// ECHO インスタンス作成
	e := echo.New()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	// ルーティングの登録
	e.GET("/", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Hello, World!"})
	})

	e.GET("/test", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message":"THIS IS TEST!!"})
	})

	if err := e.Start(fmt.Sprintf(":%d", port)); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := run(); err != nil{
		log.Fatal(err)
	}
}