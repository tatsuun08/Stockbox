package main

import (
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	// "github.com/labstack/echo/v5/middleware"
)

// 正しいパスワードで検証が成功するかテスト
func TestVerifyPassword_Success(t *testing.T) {
	password := "HelloWorld"

	// ランダムソルトの生成
	salt := make([]byte, SALTLEN)
	if _, err := rand.Read(salt); err != nil {
		t.Fatalf("ソルト生成に失敗: %v", err)
	}

	// ハッシュ化
	encodedHash, err := HashingArgon2(password, salt, MEMORY, TIME, THREADS, KEYLENGTH)
	if err != nil {
		t.Fatalf("ハッシュ化に失敗: %v", err)
	}

	// 検証実行
	ok, err := VerifyPassword(password, encodedHash)
	if err != nil {
		t.Fatalf("パスワード検証中にエラー発生: %v", err)
	}

	if !ok {
		t.Errorf("期待値: true (一致), 実際: %v (不一致)", ok)
	}
}

// 誤ったパスワードで検証が弾かれるかテスト
func TestVerifyPassword_Failure(t *testing.T) {
	password := "HelloWorld"
	wrongPassword := "WrongPassword123"

	salt := make([]byte, SALTLEN)
	if _, err := rand.Read(salt); err != nil {
		t.Fatalf("ソルト生成に失敗: %v", err)
	}

	encodedHash, err := HashingArgon2(password, salt, MEMORY, TIME, THREADS, KEYLENGTH)
	if err != nil {
		t.Fatalf("ハッシュ化に失敗: %v", err)
	}

	ok, err := VerifyPassword(wrongPassword, encodedHash)
	if err != nil {
		t.Fatalf("パスワード検証中にエラー発生: %v", err)
	}

	if ok {
		t.Errorf("期待値: false (不一致), 実際: true (誤って認証成功)")
	}
}

// CookieからのセッションIDが一致するか
func TestVerifySessionId(t *testing.T) {
	sessionId := "test_sessionID"
	e := echo.New()

	// 1. 書き込み用のリクエスト＆レスポンス
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	writeCookieSessionId(c, sessionId)

	// 2. レスポンスレコーダーから発行されたCookieを拾い上げる
	res := rec.Result()
	cookies := res.Cookies()
	var targetCookie *http.Cookie
	for _, ck := range cookies {
		if ck.Name == "sessionId" {
			targetCookie = ck
			break
		}
	}
	if targetCookie == nil {
		t.Fatalf("レスポンスにsessionIdのCookieが存在しません")
	}

	// 3. 拾ったCookieを読み取り用の新しいリクエストにセットする
	reqRead := httptest.NewRequest(http.MethodGet, "/", nil)
	reqRead.AddCookie(targetCookie) // ここでリクエストヘッダーに注入！
	recRead := httptest.NewRecorder()
	cRead := e.NewContext(reqRead, recRead)

	// 4. 読み取りテスト
	storedSessionId, err := readCookieSessionId(cRead)
	if err != nil {
		t.Fatalf("クッキーの読み取りに失敗: %v", err)
	}
	if storedSessionId != sessionId {
		t.Errorf("セッションIDが一致しません: got %v, want %v", storedSessionId, sessionId)
	}
}
