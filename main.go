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
	"stockbox/database"
)

type userRequest struct {
	Username string `json:username`
	Passwd string `json:passwd`
}

type loginRequest struct {
	Username string `json:"username"`
	Passwd string `json:"passwd"`
}

//go:embed sql/schema.sql
var ddl string

// DB接続
func connect_db() (*sql.DB, error) {
	// データベース接続
	db, err := sql.Open("sqlite", "appdb.db")
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

	// SQLC用のクエリを投げる構造体
	queries := database.New(db)

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

	e.POST("/create_user", func(c *echo.Context) error {
		user_request := new(userRequest)		

		if err := c.Bind(user_request); err != nil {
			return err
		}

		passwd := user_request.Passwd

		// DBにユーザー情報を保存
		_, err := queries.CreateUser(ctx, database.CreateUserParams{
			Username: user_request.Username,
			Passwd: passwd,
		})

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message":"Username already exists"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message":"New user created"})
	})

	e.POST("/login", func(c *echo.Context) error {
		login_request := new(loginRequest)

		// リクエストのJSONを構造体にバインド
		if err := c.Bind(login_request); err != nil {
			return err
		}
		
		// 取得したデータを使った処理をここに記述
		fmt.Println(login_request.Username)

		passwd := login_request.Passwd

		user_db, err := queries.GetUser(ctx, database.GetUserParams{
			Username: login_request.Username,
			Passwd: passwd,	
		})

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Mismatch Username or Password"})
		}

		return c.JSON(http.StatusOK, fmt.Sprintf("%s login success", user_db.Username))
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