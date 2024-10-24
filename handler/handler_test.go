package handler_test

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/ystkg/rest-example/entity"
	"github.com/ystkg/rest-example/handler"
	"github.com/ystkg/rest-example/repository"
	"github.com/ystkg/rest-example/service"
	"gopkg.in/yaml.v3"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo/v4"
)

type testPgDB struct {
	dbname string
	sqlDB  *sql.DB
	pool   *pgxpool.Pool
	conn   *pgxpool.Conn
}

var pgimage, mysqlimage *string
var pgpassword, mysqlpassword *string
var pgpool *pgxpool.Pool

func init() {
	buf, err := os.ReadFile("../docker-compose.yml")
	if err != nil {
		log.Fatal(err)
	}

	conf := struct {
		Services struct {
			Postgres struct {
				Image string
			}
			PostgresTest struct {
				Environment struct {
					PostgresPassword string `yaml:"POSTGRES_PASSWORD"`
				}
			} `yaml:"postgres-test"`
			MySQL struct {
				Image string
			}
			MySQLTest struct {
				Environment struct {
					MySQLRootPassword string `yaml:"MYSQL_ROOT_PASSWORD"`
				}
			} `yaml:"mysql-test"`
		}
	}{}
	if err = yaml.Unmarshal(buf, &conf); err != nil {
		log.Fatal(err)
	}

	pgimage = &conf.Services.Postgres.Image
	mysqlimage = &conf.Services.MySQL.Image
	pgpassword = &conf.Services.PostgresTest.Environment.PostgresPassword
	mysqlpassword = &conf.Services.MySQLTest.Environment.MySQLRootPassword

	if pgpool, err = pgxpool.New(context.Background(), formatDSN("postgres")); err != nil {
		log.Fatal(err)
	}
}

func formatDSN(user string) string {
	// dbnameは指定しないとuserと同じになる
	return fmt.Sprintf("host=localhost port=15432 user=%s password=%s sslmode=disable TimeZone=Asia/Tokyo", user, *pgpassword)
}

func createTestDatabase(dbname string) (string, error) {
	ctx := context.Background()

	conn, err := pgpool.Acquire(ctx)
	if err != nil {
		return "", err
	}
	defer conn.Release()

	if err := dropDatabaseIfExists(ctx, conn, dbname); err != nil {
		return "", err
	}

	// dbnameと同一名のuser（LOGIN権限のあるROLE）を作成
	if _, err := conn.Exec(ctx, "CREATE ROLE "+dbname+" LOGIN PASSWORD '"+*pgpassword+"'"); err != nil {
		return "", err
	}

	if _, err := conn.Exec(ctx, "CREATE DATABASE "+dbname+" OWNER "+dbname); err != nil {
		return "", err
	}

	return formatDSN(dbname), nil
}

func dropDatabaseIfExists(ctx context.Context, conn *pgxpool.Conn, dbname string) error {
	if _, err := conn.Exec(ctx, "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1", dbname); err != nil {
		return err
	}

	batch := &pgx.Batch{}
	batch.Queue("DROP DATABASE IF EXISTS " + dbname)
	batch.Queue("DROP ROLE IF EXISTS " + dbname)
	results := conn.SendBatch(ctx, batch)
	defer results.Close()
	for range batch.Len() {
		if _, err := results.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func setupTest(testname string) (*echo.Echo, *handler.HandlerConfig, *testPgDB, pgx.Tx, error) {
	e, conf, testDB, tx, _, err := setupTestMain(testname, service.NewService)
	if err != nil {
		cleanDB(testDB)
		testDB = nil
	}
	return e, conf, testDB, tx, err
}

func setupMockTest(testname string) (*echo.Echo, *handler.HandlerConfig, *testPgDB, pgx.Tx, *serviceMock, error) {
	newService := func(r repository.Repository) service.Service {
		return newMockService(newMockRepository(r))
	}
	e, conf, testDB, tx, s, err := setupTestMain(testname, newService)
	if err != nil {
		cleanDB(testDB)
		testDB = nil
	}
	var mock *serviceMock
	if s != nil {
		mock = s.(*serviceMock)
	}
	return e, conf, testDB, tx, mock, err
}

func setupTestMain(testname string, newService func(repository.Repository) service.Service) (*echo.Echo, *handler.HandlerConfig, *testPgDB, pgx.Tx, service.Service, error) {
	ctx := context.Background()

	// Database
	dbname := strings.ToLower(testname)
	dburl, err := createTestDatabase(dbname)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	// Repository
	pool, err := pgxpool.New(ctx, dburl)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	sqlDB := stdlib.OpenDBFromPool(pool)
	testDB := &testPgDB{dbname: dbname, sqlDB: sqlDB, pool: pool}
	driverName := "pgx"
	r, err := repository.NewRepository(driverName, sqlDB)
	if err != nil {
		return nil, nil, testDB, nil, nil, err
	}
	if err := r.InitDb(ctx); err != nil {
		return nil, nil, testDB, nil, nil, err
	}

	// Service
	s := newService(r)

	// Handler
	jwtkey := []byte(testname)
	validityMin := 1
	location, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return nil, nil, testDB, nil, nil, err
	}
	conf := &handler.HandlerConfig{
		JwtKey:           jwtkey,
		ValidityMin:      validityMin,
		DateTimeLayout:   time.DateTime,
		Location:         location,
		Indent:           "  ",
		TimeoutSec:       60,
		RequestBodyLimit: "1K",
		RateLimit:        10,
	}
	h := handler.NewHandler(s, conf)

	// Echo
	e := handler.NewEcho(h)

	// トランザクション
	if testDB.conn, err = pool.Acquire(ctx); err != nil {
		return nil, nil, testDB, nil, nil, err
	}
	tx, err := testDB.conn.Begin(ctx)
	if err != nil {
		return nil, nil, testDB, nil, nil, err
	}

	return e, conf, testDB, tx, s, nil
}

func setupMySQLTest(testname string) (*echo.Echo, error) {
	ctx := context.Background()

	// Repository
	driverName := "mysql"
	sqlDB, err := sql.Open(driverName, fmt.Sprintf("root:%s@tcp(localhost:13306)/testdb?parseTime=true", *mysqlpassword))
	if err != nil {
		return nil, err
	}
	r, err := repository.NewRepository(driverName, sqlDB)
	if err != nil {
		sqlDB.Close()
		return nil, err
	}
	if err := r.InitDb(ctx); err != nil {
		sqlDB.Close()
		return nil, err
	}
	sqlDB.ExecContext(ctx, "DELETE FROM users WHERE name = ?", testname)

	// Service
	s := service.NewService(r)

	// Handler
	jwtkey := []byte(testname)
	validityMin := 1
	location, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		sqlDB.Close()
		return nil, err
	}
	conf := &handler.HandlerConfig{
		JwtKey:           jwtkey,
		ValidityMin:      validityMin,
		DateTimeLayout:   time.DateTime,
		Location:         location,
		Indent:           "  ",
		TimeoutSec:       60,
		RequestBodyLimit: "1K",
		RateLimit:        10,
	}
	h := handler.NewHandler(s, conf)

	// Echo
	e := handler.NewEcho(h)

	return e, nil
}

func cleanIfSuccess(t *testing.T, testDB *testPgDB) error {
	if t.Failed() {
		testDB.sqlDB.Close()
		return nil
	}
	return cleanDB(testDB)
}

func cleanDB(testDB *testPgDB) error {
	if testDB == nil {
		return nil
	}
	if testDB.sqlDB == nil {
		return nil
	}
	testDB.sqlDB.Close()

	ctx := context.Background()

	conn, err := pgpool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	return dropDatabaseIfExists(ctx, conn, testDB.dbname)
}

func execHandlerTest(e *echo.Echo, testDB *testPgDB, tx pgx.Tx, req *http.Request) (*httptest.ResponseRecorder, *DiffTables, *TestDB, error) {
	loadedTime := time.Now()

	ctx := context.Background()

	tables, err := loadTables(tx)
	if err != nil {
		tx.Rollback(ctx)
		return nil, nil, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		return nil, nil, nil, err
	}

	testDB.conn.Release()

	before := &TestDB{
		loadedTime,
		tables,
	}

	rec, err := execHandler(e, req)
	if err != nil {
		return nil, nil, nil, err
	}

	conn, err := testDB.pool.Acquire(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	defer conn.Release()

	tx, err = conn.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, nil, nil, err
	}
	defer tx.Rollback(ctx)

	loadedTime = time.Now()

	tables, err = loadTables(tx)
	if err != nil {
		return nil, nil, nil, err
	}

	after := &TestDB{
		loadedTime,
		tables,
	}

	return rec, diffTables(before, after), before, nil
}

func execHandlerValidation(e *echo.Echo, req *http.Request) (int, error, error) {
	rec, err := execHandler(e, req)
	if err == nil {
		return rec.Code, nil, nil
	}

	httpError, ok := err.(*echo.HTTPError)
	if !ok {
		return 0, nil, err
	}

	return httpError.Code, httpError.Internal.(interface{ Unwrap() error }).Unwrap(), nil
}

func execHandler(e *echo.Echo, req *http.Request) (*httptest.ResponseRecorder, error) {
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Router().Find(req.Method, req.URL.Path, c)
	return rec, c.Handler()(c)
}

func newRequest(method, target string, body *string, contentType string, jwt *string) *http.Request {
	var reader io.Reader
	if body != nil {
		reader = strings.NewReader(*body)
	}

	req := httptest.NewRequest(method, target, reader)

	if contentType != "" {
		req.Header.Set(echo.HeaderContentType, contentType)
	}

	if jwt != nil {
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+*jwt)
	}

	return req
}

func genToken(conf *handler.HandlerConfig, userId uint) *string {
	iat := time.Now()
	claims := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&handler.JwtCustomClaims{
			jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(iat.Add(time.Duration(conf.ValidityMin) * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(iat),
			},
			userId,
		},
	)

	signed, err := claims.SignedString(conf.JwtKey)
	if err != nil {
		log.Fatal(err)
	}

	return &signed
}

type Tables struct {
	users  map[uint]*entity.User
	prices map[uint]*entity.Price
}

func (t *Tables) count() (c int) {
	c = 0
	if len(t.users) != 0 {
		c += 1
	}
	if len(t.prices) != 0 {
		c += 1
	}
	return
}

// X:instead of 'lint:ignore U1000'. remove X if use.
func (t *Tables) XfindUser(ID uint) *entity.User {
	entity, ok := t.users[ID]
	if !ok {
		return nil
	}
	return entity
}

func (t *Tables) userAny() *entity.User {
	for _, v := range t.users {
		return v
	}
	return nil
}

// X:instead of 'lint:ignore U1000'. remove X if use.
func (t *Tables) XuserList() []*entity.User {
	entities := make([]*entity.User, 0, len(t.users))
	for _, v := range t.users {
		entities = append(entities, v)
	}
	sort.Slice(entities, func(i, j int) bool { return entities[i].ID < entities[j].ID }) // asc
	return entities
}

func (t *Tables) findPrice(ID uint) *entity.Price {
	entity, ok := t.prices[ID]
	if !ok {
		return nil
	}
	return entity
}

func (t *Tables) priceAny() *entity.Price {
	for _, v := range t.prices {
		return v
	}
	return nil
}

// X:instead of 'lint:ignore U1000'. remove X if use.
func (t *Tables) XpriceList() []*entity.Price {
	entities := make([]*entity.Price, 0, len(t.prices))
	for _, v := range t.prices {
		entities = append(entities, v)
	}
	sort.Slice(entities, func(i, j int) bool { return entities[i].ID < entities[j].ID }) // asc
	return entities
}

type TestDB struct {
	loadedTime time.Time
	*Tables
}

func loadTables(tx pgx.Tx) (*Tables, error) {
	users, err := loadUsers(tx)
	if err != nil {
		return nil, err
	}

	prices, err := loadPrices(tx)
	if err != nil {
		return nil, err
	}

	return &Tables{
			users:  users,
			prices: prices,
		},
		nil
}

func loadUsers(tx pgx.Tx) (map[uint]*entity.User, error) {
	const SQL = "SELECT id, created_at, updated_at, deleted_at, name, password FROM users"
	rows, err := tx.Query(context.Background(), SQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make(map[uint]*entity.User)
	for rows.Next() {
		user := &entity.User{}
		rows.Scan(
			&user.ID,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
			&user.Name,
			&user.Password,
		)
		users[user.ID] = user
	}

	return users, nil
}

func loadPrices(tx pgx.Tx) (map[uint]*entity.Price, error) {
	const SQL = "SELECT id, created_at, updated_at, deleted_at, user_id, date_time, store, product, price FROM prices"
	rows, err := tx.Query(context.Background(), SQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prices := make(map[uint]*entity.Price)
	for rows.Next() {
		price := &entity.Price{}
		rows.Scan(
			&price.ID,
			&price.CreatedAt,
			&price.UpdatedAt,
			&price.DeletedAt,
			&price.UserID,
			&price.DateTime,
			&price.Store,
			&price.Product,
			&price.Price,
		)
		prices[price.ID] = price
	}

	return prices, nil
}

type DiffTables struct {
	created         *Tables
	updated         *Tables
	logicalDeleted  *Tables
	physicalDeleted *Tables
}

func diffTables(before, after *TestDB) *DiffTables {
	createdUsers, updatedUsers, logicalDeletedUsers, physicalDeletedUsers := diffUser(before.users, after.users)
	createdPrices, updatedPrices, logicalDeletedPrices, physicalDeletedPrices := diffPrice(before.prices, after.prices)

	diff := &DiffTables{
		created: &Tables{
			users:  createdUsers,
			prices: createdPrices,
		},
		updated: &Tables{
			users:  updatedUsers,
			prices: updatedPrices,
		},
		logicalDeleted: &Tables{
			users:  logicalDeletedUsers,
			prices: logicalDeletedPrices,
		},
		physicalDeleted: &Tables{
			users:  physicalDeletedUsers,
			prices: physicalDeletedPrices,
		},
	}

	if diff.created.count() == 0 && diff.updated.count() == 0 && diff.logicalDeleted.count() == 0 && diff.physicalDeleted.count() == 0 {
		return nil
	}

	return diff
}

func diffUser(before, after map[uint]*entity.User) (created, updated, logical, physical map[uint]*entity.User) {
	created = make(map[uint]*entity.User, 0)
	updated = make(map[uint]*entity.User, 0)
	logical = make(map[uint]*entity.User, 0)
	physical = make(map[uint]*entity.User, 0)
	for k, v := range after {
		before, ok := before[k]
		if !ok {
			created[k] = v
		} else if v.DeletedAt.Valid && !before.DeletedAt.Valid {
			logical[k] = v
		} else if *v != *before {
			updated[k] = v
		}
	}
	for k, v := range before {
		if _, ok := after[k]; !ok {
			physical[k] = v
		}
	}
	return
}

func diffPrice(before, after map[uint]*entity.Price) (created, updated, logical, physical map[uint]*entity.Price) {
	created = make(map[uint]*entity.Price, 0)
	updated = make(map[uint]*entity.Price, 0)
	logical = make(map[uint]*entity.Price, 0)
	physical = make(map[uint]*entity.Price, 0)
	for k, v := range after {
		before, ok := before[k]
		if !ok {
			created[k] = v
		} else if v.DeletedAt.Valid && !before.DeletedAt.Valid {
			logical[k] = v
		} else if *v != *before {
			updated[k] = v
		}
	}
	for k, v := range before {
		if _, ok := after[k]; !ok {
			physical[k] = v
		}
	}
	return
}

func insertUser(tx pgx.Tx, createdAt, updatedAt, deletedAt *time.Time, name, password string) (id uint, err error) {
	const SQL = "INSERT INTO users (created_at, updated_at, deleted_at, name, password) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	err = tx.QueryRow(context.Background(), SQL, createdAt, updatedAt, deletedAt, name, password).Scan(&id)
	return
}

func insertUsers(tx pgx.Tx, t *time.Time, rows [][2]any) (int64, error) {
	inputRows := make([][]any, len(rows))
	for i, v := range rows {
		inputRows[i] = []any{t, t, v[0], v[1]}
	}
	return tx.CopyFrom(
		context.Background(),
		pgx.Identifier{"users"},
		[]string{"created_at", "updated_at", "name", "password"},
		pgx.CopyFromRows(inputRows),
	)
}

func insertPrice(tx pgx.Tx, createdAt, updatedAt, deletedAt *time.Time, userID uint, dateTime time.Time, store, product string, price uint) (id uint, err error) {
	const SQL = "INSERT INTO prices (created_at, updated_at, deleted_at, user_id, date_time, store, product, price) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id"
	err = tx.QueryRow(context.Background(), SQL, createdAt, updatedAt, deletedAt, userID, dateTime, store, product, price).Scan(&id)
	return
}

func insertPrices(tx pgx.Tx, t *time.Time, rows [][5]any) (int64, error) {
	inputRows := make([][]any, len(rows))
	for i, v := range rows {
		inputRows[i] = []any{t, t, v[0], v[1], v[2], v[3], v[4]}
	}
	return tx.CopyFrom(
		context.Background(),
		pgx.Identifier{"prices"},
		[]string{"created_at", "updated_at", "user_id", "date_time", "store", "product", "price"},
		pgx.CopyFromRows(inputRows),
	)
}
