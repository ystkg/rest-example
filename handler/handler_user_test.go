package handler_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ystkg/rest-example/api"
	"github.com/ystkg/rest-example/handler"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func someUsers() [][2]any {
	return [][2]any{
		{"testuser1", "testpassword1"},
		{"testuser2", "testpassword2"},
	}
}

// ユーザの登録の正常系
func TestCreateUser(t *testing.T) {
	testname := "TestCreateUser"

	// セットアップ
	e, _, testDB, tx, err := setupTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(t, testDB)

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
	rec, diff, _, err := execHandlerTest(e, testDB, tx, req)
	if err != nil {
		t.Fatal(err)
	}

	// アサーション
	assert.Equal(t, 201, rec.Code)

	res := &api.User{}
	if err := json.Unmarshal(rec.Body.Bytes(), res); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, name, res.Name)

	assert.NotNil(t, diff)
	assert.Equal(t, 1, diff.created.count())
	assert.Zero(t, diff.updated.count())
	assert.Zero(t, diff.logicalDeleted.count())
	assert.Zero(t, diff.physicalDeleted.count())

	assert.Equal(t, 1, len(diff.created.users))
	entity := diff.created.userAny() // just one
	assert.Equal(t, *res.ID, entity.ID)
	assert.False(t, entity.DeletedAt.Valid)
	assert.Equal(t, name, entity.Name)
	assert.Equal(t, encodePassword(name, password), entity.Password)
}

// ユーザの登録のバリデーション
func TestCreateUserValidation(t *testing.T) {
	testname := "TestCreateUserValidation"

	// セットアップ
	e, _, testDB, tx, err := setupTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(t, testDB)

	// バリデーションのテストは事前にコミットしてテーブル駆動
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		body string
		code int
		err  error
	}{
		{"name=testuser01&password=testpassword", 201, nil},
		{"name=testuser01&password=testpassword", 400, handler.ErrAlreadyRegistered},
		{"", 400, nil},
		{"name=testuser02", 400, nil},
		{"password=testpassword", 400, nil},
		{"name=_testuser01&password=testpassword", 400, nil},
		{"name=%0g", 400, nil},
		{"name=testuser02&password=pw", 400, nil},
	}

	for _, v := range cases {
		// リクエストの生成
		req := newRequest(
			http.MethodPost,
			"/users",
			&v.body,
			echo.MIMEApplicationForm,
			nil,
		)

		// テストの実行
		code, cause, err := execHandlerValidation(e, req)
		if err != nil {
			t.Fatal(err)
		}

		// アサーション
		assert.Equal(t, v.code, code)
		if v.err != nil {
			assert.Equal(t, v.err, cause)
		}
	}
}

// トークン発行の正常系
func TestGenToken(t *testing.T) {
	testname := "TestGenToken"

	// セットアップ
	e, conf, testDB, tx, err := setupTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(t, testDB)

	// データベースの初期データ生成
	now := time.Now()
	if _, err := insertUsers(tx, &now, someUsers()); err != nil {
		t.Fatal(err)
	}

	name, password := "testuser01", "testpassword"
	id, err := insertUser(tx, &now, &now, nil, name, encodePassword(name, password))
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
	rec, diff, _, err := execHandlerTest(e, testDB, tx, req)
	if err != nil {
		t.Fatal(err)
	}

	// アサーション
	assert.Equal(t, 201, rec.Code)

	res := &api.UserToken{}
	if err := json.Unmarshal(rec.Body.Bytes(), res); err != nil {
		t.Fatal(err)
	}

	token := strings.Split(res.Token, ".")
	assert.Equal(t, 3, len(token))
	payload := token[1]
	header, claims, signature, err := decodeJwt(payload, conf.JwtKey)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, header+"."+payload+"."+signature, res.Token)
	assert.Equal(t, id, claims.UserId)

	assert.Nil(t, diff)
}

// トークン発行のバリデーション
func TestGenTokenValidation(t *testing.T) {
	testname := "TestGenTokenValidation"

	// セットアップ
	e, _, testDB, tx, err := setupTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(t, testDB)

	// データベースの初期データ生成
	name, password := "testuser01", "testpassword"
	now := time.Now()
	if _, err := insertUser(tx, &now, &now, nil, name, encodePassword(name, password)); err != nil {
		t.Fatal(err)
	}

	// バリデーションのテストは事前にコミットしてテーブル駆動
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		body string
		code int
		err  error
	}{
		{name, "password=" + password, 201, nil},
		{"", "", 404, handler.ErrNotFound},
		{name, "password=" + password + "a", 401, handler.ErrAuthenticationFailed},
		{name + "a", "password=" + password, 401, handler.ErrAuthenticationFailed},
	}

	for _, v := range cases {
		// リクエストの生成
		req := newRequest(
			http.MethodPost,
			"/users/"+v.name+"/token",
			&v.body,
			echo.MIMEApplicationForm,
			nil,
		)

		// テストの実行
		code, cause, err := execHandlerValidation(e, req)
		if err != nil {
			t.Fatal(err)
		}

		// アサーション
		assert.Equal(t, v.code, code)
		if v.err != nil {
			assert.Equal(t, v.err, cause)
		}
	}
}

func encodePassword(name, password string) string {
	sha256 := sha256.Sum256([]byte(fmt.Sprintf("%s %s", name, password)))
	return hex.EncodeToString(sha256[:])
}

func decodeJwt(payload string, jwtkey []byte) (string, *handler.JwtCustomClaims, string, error) {
	header := base64.RawStdEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	bytes, err := base64.RawStdEncoding.DecodeString(payload)
	if err != nil {
		return "", nil, "", err
	}

	claims := &handler.JwtCustomClaims{}
	if err := json.Unmarshal(bytes, claims); err != nil {
		return "", nil, "", err
	}

	signature, err := jwt.SigningMethodHS256.Sign(header+"."+payload, jwtkey)
	if err != nil {
		return "", nil, "", err
	}

	return header, claims, base64.RawURLEncoding.EncodeToString(signature), nil
}
