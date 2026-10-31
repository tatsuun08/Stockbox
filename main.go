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
	"time"
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
	MEMORY = 64 * 1024
	TIME = 3
	THREADS = 4
	KEYLENGTH = 32
	SALTLEN = 16
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

// セッションIDの作成
func generateSessionID() (string, error) {
	// 32バイト（256ビット）の十分なエントロピーを確保
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// CookieやURLに入れても安全なBase64文字列（RawURLEncoding: パディングなし）に変換
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// セッションIDをCookieに書き込み
func writeCookieSessionId(c *echo.Context, sessionId string) {
	cookie := new(http.Cookie)
	cookie.Name = "sessionId"
	cookie.Value = sessionId
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.Path = "/"
	c.SetCookie(cookie)
}

// セッションIDをCookieから探す
func readCookieSessionId(c *echo.Context) (string, error) {
	cookie, err := c.Cookie("sessionId")
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}

// セッション確認用のミドルウェア
func SessionMiddleware(queries *database.Queries) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			ctx := c.Request().Context()
			
			// クッキーを読み込む
			sessionId, err := readCookieSessionId(c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message":"クッキーが見つかりません"})
			}
			
			// DBで探す
			loginSession, err := queries.GetLoginSession(ctx, sessionId)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message":"ログインセッションが見つかりません"})
			}

			// 期限切れチェック
			if !loginSession.ExpiresAt.Valid {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message":"期限情報なし"})
			}

			expiresAt, err := time.Parse("2006-01-02 15:04:05", loginSession.ExpiresAt.String)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"message":"internal server error"})
			}

			t := time.Now()
			
			// 切れていたら
			if t.After(expiresAt) {
				// セッションの削除
				_ = queries.DeleteLoginSession(ctx, sessionId)
				return c.JSON(http.StatusUnauthorized, map[string]string{"message":"セッション期限切れです"})
			}
			// OKなら
			c.Set("user_id", loginSession.UserID)
			c.Set("session_id", sessionId)
			return next(c)
		}
	}
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

	e.POST("/create_user", func(c *echo.Context) error {
		user_request := new(userRequest)		

		if err := c.Bind(user_request); err != nil {
			return err
		}
		// ソルトの作成
		salt := make([]byte, SALTLEN)
		if _, err := rand.Read(salt); err != nil {
			return err
		}
		// パスワードのハッシュ化
		passwd, err := HashingArgon2(user_request.Passwd, salt, MEMORY, TIME, THREADS, KEYLENGTH)
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
			// セッションを検索
			_, err := readCookieSessionId(c)
			if err != nil {
				// セッションの追加
				sessionId, err := generateSessionID()
				if err != nil {
					return err
				}
				// クッキーへの書き込み
				writeCookieSessionId(c, sessionId)

				// DBへの保存
				_, err = queries.CreateSession(
					ctx,
					database.CreateSessionParams{
						SessionID: sessionId,
						UserID: user_db.ID,
					},
				)
				if err != nil {
					return err
				}
			}
			return c.JSON(http.StatusOK, fmt.Sprintf("%s login success", user_db.Username))
		}else {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Mismatch Username or Password"})
		}

	})

	// ミドルウェアによる検問
	auth := e.Group("")
	auth.Use(SessionMiddleware(queries))

	// read box 自身のストックボックス一覧 (50)
	e.POST("/my-box", func(c *echo.Context) error {


		return nil
	})

	// create box

	// delete box

	// create read_later
	
	// delete read_later

	// edit read_later 

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