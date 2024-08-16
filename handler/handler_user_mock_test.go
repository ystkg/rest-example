package handler_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestCreateUserCreateError(t *testing.T) {
	testname := "TestCreateUserCreateError"

	// セットアップ
	e, sqlDB, mock, tx, _, _, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t, sqlDB)

	// mockの挙動設定
	mock.repository.user.err = errors.New(testname)

	// データベースの初期データ生成
	now := time.Now()
	if _, err := insertUsers(tx, &now, someUsers()); err != nil {
		t.Fatal(err)
	}

	// リクエストの生成
	name, password := "testuser01", "testpassword"
	body := fmt.Sprintf("name=%s&password=%s", name, password)
	req := newRequest(
		http.MethodPost,
		"/users",
		&body,
		echo.MIMEApplicationForm,
		nil,
	)

	// テストの実行
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.repository.user.err, err)
	assert.Nil(t, diff)
}

func TestCreateUserBeginTxError(t *testing.T) {
	testname := "TestCreateUserBeginTxError"

	// セットアップ
	e, sqlDB, mock, tx, _, _, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t, sqlDB)

	// mockの挙動設定
	mock.repository.beginTxErr = errors.New(testname)

	// データベースの初期データ生成
	now := time.Now()
	if _, err := insertUsers(tx, &now, someUsers()); err != nil {
		t.Fatal(err)
	}

	// リクエストの生成
	name, password := "testuser01", "testpassword"
	body := fmt.Sprintf("name=%s&password=%s", name, password)
	req := newRequest(
		http.MethodPost,
		"/users",
		&body,
		echo.MIMEApplicationForm,
		nil,
	)

	// テストの実行
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.repository.beginTxErr, err)
	assert.Nil(t, diff)
}

func TestCreateUserCommitError(t *testing.T) {
	testname := "TestCreateUserCommitError"

	// セットアップ
	e, sqlDB, mock, tx, _, _, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t, sqlDB)

	// mockの挙動設定
	mock.repository.commitErr = errors.New(testname)

	// データベースの初期データ生成
	now := time.Now()
	if _, err := insertUsers(tx, &now, someUsers()); err != nil {
		t.Fatal(err)
	}

	// リクエストの生成
	name, password := "testuser01", "testpassword"
	body := fmt.Sprintf("name=%s&password=%s", name, password)
	req := newRequest(
		http.MethodPost,
		"/users",
		&body,
		echo.MIMEApplicationForm,
		nil,
	)

	// テストの実行
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.repository.commitErr, err)
	assert.Nil(t, diff)
}

func TestGenTokenFindError(t *testing.T) {
	testname := "TestGenTokenFindError"

	// セットアップ
	e, sqlDB, mock, tx, _, _, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t, sqlDB)

	// mockの挙動設定
	mock.repository.user.err = errors.New(testname)

	// データベースの初期データ生成
	now := time.Now()
	if _, err := insertUsers(tx, &now, someUsers()); err != nil {
		t.Fatal(err)
	}

	name, password := "testuser01", "testpassword"
	_, err = insertUser(tx, &now, &now, nil, name, encodePassword(name, password))
	if err != nil {
		t.Fatal(err)
	}

	// リクエストの生成
	body := "password=" + password
	req := newRequest(
		http.MethodPost,
		"/users/"+name+"/token",
		&body,
		echo.MIMEApplicationForm,
		nil,
	)

	// テストの実行
	_, diff, _, err := execHandlerTest(e, tx, req)

	// アサーション
	assert.Equal(t, mock.repository.user.err, err)
	assert.Nil(t, diff)
}
