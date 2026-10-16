package main

import (
	"context"
	// "crypto/rand"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	_ "embed"
	"encoding/base64"

	// "encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"golang.org/x/crypto/argon2"

	"stockbox/database"
)

const (
	memory = 64 * 1024
	time = 3
	threads = 4
	keyLength = 32
	saltLen = 16
)

type userRequest struct {
	Username string `json:"username"`
	Passwd string `json:"passwd"`
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


// Hashing with argon2
func HashingArgon2(password string, salt []byte, time uint32, memory uint32, threads uint8, keyLen uint32)(string, error){
	// テキストのBase64エンコーディング
	encodedPassword := []byte(password)

	hash := argon2.IDKey(encodedPassword, salt, time, memory, threads, keyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt) // (byte -> b64 string)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash) // (byte -> b64 string)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, memory, time , threads, b64Salt, b64Hash), nil
}

// test argon2
func VerifyPassword(password string, encodedHash string)(bool, error) {
	vals := strings.Split(encodedHash, "$")
	if (len(vals) != 6){
		return false, errors.New("invalid hash format")
	}

	// パラメータの取得　メモリ、　時間、並列度
	var timeCost, memoryCost uint32
	var threads uint8
	_, err := fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &memoryCost, &timeCost, &threads)

	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return false, errors.New("failed to decode salt")
	}
	// ハッシュ値の取得
	storedHash, err := base64.RawStdEncoding.DecodeString(vals[5])
	if err != nil {
		return false, errors.New("failed to decode hash")
	}
	// 検出したソルトでパスワードをハッシュ化（Argon2）
	keyLen := len(storedHash)

	comparisonHash := argon2.IDKey([]byte(password), salt, timeCost, memoryCost, uint8(threads), uint32(keyLen))

	if subtle.ConstantTimeCompare(comparisonHash, storedHash) == 1 {
		return true, nil
	}

	return false, nil
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
		// ソルトの作成
		salt := make([]byte, saltLen)
		if _, err := rand.Read(salt); err != nil {
			return err
		}
		// パスワードのハッシュ化
		passwd, err := HashingArgon2(user_request.Passwd, salt, memory, time, threads, keyLength)
		if err != nil {
			return err
		}


		// DBにユーザー情報を保存
		if _, err := queries.CreateUser(ctx, database.CreateUserParams{
			Username: user_request.Username,
			Passwd: passwd,
		}); err != nil {
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

		passwd := login_request.Passwd

		user_db, err := queries.GetUser(ctx, login_request.Username)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Mismatch Username or Password"})
		}

		var isVerified bool

		if isVerified, err = VerifyPassword(passwd, user_db.Passwd); err != nil {
			return err
		}

		if isVerified {
			return c.JSON(http.StatusOK, fmt.Sprintf("%s login success", user_db.Username))
		}else {
			return c.JSON(http.StatusInternalServerError, "login failed")
		}

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