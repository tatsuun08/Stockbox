package main

import (
	"crypto/rand"
	"testing"
)

// 正しいパスワードで検証が成功するかテスト
func TestVerifyPassword_Success(t *testing.T) {
	password := "HelloWorld"

	// ランダムソルトの生成
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		t.Fatalf("ソルト生成に失敗: %v", err)
	}

	// ハッシュ化
	encodedHash, err := HashingArgon2(password, salt, memory, time, threads, keyLength)
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

	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		t.Fatalf("ソルト生成に失敗: %v", err)
	}

	encodedHash, err := HashingArgon2(password, salt, memory, time, threads, keyLength)
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