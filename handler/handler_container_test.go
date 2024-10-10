package handler_test

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/labstack/echo/v4"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/ystkg/rest-example/handler"
	"github.com/ystkg/rest-example/repository"
	"github.com/ystkg/rest-example/service"
)

func setupContainerTest(testname, driverName string, sqlDB *sql.DB) (*echo.Echo, error) {
	// Repository
	r, err := repository.NewRepository(driverName, sqlDB)
	if err != nil {
		return nil, err
	}
	if err := r.InitDb(context.Background()); err != nil {
		return nil, err
	}

	// Service
	s := service.NewService(r)

	// Handler
	location, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return nil, err
	}
	conf := &handler.HandlerConfig{
		JwtKey:         []byte(testname),
		ValidityMin:    1,
		DateTimeLayout: time.DateTime,
		Location:       location,
		Indent:         "  ",
		TimeoutSec:     60,
	}
	h := handler.NewHandler(s, conf)

	// Echo
	e := handler.NewEcho(h)

	return e, nil
}

// Testcontainers（PostgreSQL）
func TestContainerPg(t *testing.T) {
	testname := "TestContainerPg"

	if testing.Short() { // go test -short
		t.SkipNow()
	}

	password := *pgpassword

	// コンテナ起動
	ctx := context.Background()
	pgContainer, err := postgres.Run(ctx,
		*pgimage,
		postgres.WithPassword(password),
		postgres.BasicWaitStrategies(),
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_INITDB_ARGS": "--no-locale -E UTF-8",
			"TZ":                   "Asia/Tokyo",
		}),
	)
	require.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	// セットアップ
	driverName := "pgx"
	sqlDB, err := sql.Open(driverName, pgContainer.MustConnectionString(ctx, "sslmode=disable", "TimeZone=Asia/Tokyo"))
	require.NoError(t, err)
	defer sqlDB.Close()
	e, err := setupContainerTest(testname, driverName, sqlDB)
	require.NoError(t, err)

	// テストシナリオの実行
	scenario(t, e, testname)
}

// Testcontainers（MySQL）
func TestContainerMySQL(t *testing.T) {
	testname := "TestContainerMySQL"

	if testing.Short() { // go test -short
		t.SkipNow()
	}

	password := *mysqlpassword

	// コンテナ起動
	ctx := context.Background()
	driverName := "mysql"
	mysqlContainer, err := mysql.Run(ctx,
		*mysqlimage,
		mysql.WithUsername("root"),
		mysql.WithPassword(password),
		mysql.WithDatabase("testdb"),
		testcontainers.WithEnv(map[string]string{
			"TZ": "Asia/Tokyo",
		}),
		testcontainers.WithWaitStrategy(
			wait.ForSQL("3306", driverName,
				func(host string, port nat.Port) string {
					return fmt.Sprintf("root:%s@tcp(%s)/testdb", password, net.JoinHostPort(host, port.Port()))
				},
			).WithStartupTimeout(30*time.Second).WithPollInterval(1*time.Second),
		),
	)
	require.NoError(t, err)
	defer mysqlContainer.Terminate(ctx)

	// セットアップ
	sqlDB, err := sql.Open(driverName, mysqlContainer.MustConnectionString(ctx, "parseTime=true"))
	require.NoError(t, err)
	defer sqlDB.Close()
	e, err := setupContainerTest(testname, driverName, sqlDB)
	require.NoError(t, err)

	// テストシナリオの実行
	scenario(t, e, testname)
}

// Dockertest（PostgreSQL）
func TestDockerPg(t *testing.T) {
	testname := "TestDockerPg"

	if testing.Short() { // go test -short
		t.SkipNow()
	}

	image := strings.SplitN(*pgimage, ":", 2)
	repository := image[0]
	var tag string
	if len(image) == 2 {
		tag = image[1]
	}
	password := *pgpassword

	// コンテナ起動
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)
	err = pool.Client.Ping()
	require.NoError(t, err)
	resource, err := pool.Run(repository, tag, []string{
		"POSTGRES_PASSWORD=" + password,
		"POSTGRES_INITDB_ARGS=--no-locale -E UTF-8",
		"TZ=Asia/Tokyo",
	})
	require.NoError(t, err)
	defer pool.Purge(resource)

	// データベース起動完了待ち
	driverName := "pgx"
	sqlDB, err := sql.Open(driverName,
		fmt.Sprintf("postgres://postgres:%s@localhost:%s/postgres?sslmode=disable&TimeZone=Asia/Tokyo",
			password,
			resource.GetPort("5432/tcp"),
		))
	require.NoError(t, err)
	defer sqlDB.Close()
	err = pool.Retry(func() error {
		return sqlDB.Ping()
	})
	require.NoError(t, err)

	// セットアップ
	e, err := setupContainerTest(testname, driverName, sqlDB)
	require.NoError(t, err)

	// テストシナリオの実行
	scenario(t, e, testname)
}

// Dockertest（MySQL）
func TestDockerMySQL(t *testing.T) {
	testname := "TestDockerMySQL"

	if testing.Short() { // go test -short
		t.SkipNow()
	}

	image := strings.SplitN(*mysqlimage, ":", 2)
	repository := image[0]
	var tag string
	if len(image) == 2 {
		tag = image[1]
	}
	password := *mysqlpassword

	// コンテナ起動
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)
	err = pool.Client.Ping()
	require.NoError(t, err)
	resource, err := pool.Run(repository, tag, []string{
		"MYSQL_ROOT_PASSWORD=" + password,
		"MYSQL_DATABASE=testdb",
		"TZ=Asia/Tokyo",
	})
	require.NoError(t, err)
	defer pool.Purge(resource)

	// データベース起動完了待ち
	driverName := "mysql"
	sqlDB, err := sql.Open(driverName,
		fmt.Sprintf("root:%s@tcp(localhost:%s)/testdb?parseTime=true",
			password,
			resource.GetPort("3306/tcp"),
		))
	require.NoError(t, err)
	defer sqlDB.Close()
	pool.MaxWait = 30 * time.Second
	err = pool.Retry(func() error {
		return sqlDB.Ping()
	})
	require.NoError(t, err)

	// セットアップ
	e, err := setupContainerTest(testname, driverName, sqlDB)
	require.NoError(t, err)

	// テストシナリオの実行
	scenario(t, e, testname)
}
