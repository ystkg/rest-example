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
	"github.com/labstack/echo/v4"
)

var pgpassword *string

func init() {
	buf, err := os.ReadFile("../docker-compose.yml")
	if err != nil {
		log.Fatal(err)
	}

	conf := struct {
		Services struct {
			PostgresTest struct {
				Environment struct {
					PostgresPassword string `yaml:"POSTGRES_PASSWORD"`
				}
			} `yaml:"postgres-test"`
		}
	}{}
	if err = yaml.Unmarshal(buf, &conf); err != nil {
		log.Fatal(err)
	}

	pgpassword = &conf.Services.PostgresTest.Environment.PostgresPassword
}

func formatDSN(dbname string) string {
	dbnameAttr := ""
	if dbname != "" {
		dbnameAttr = "dbname=" + dbname
	}
	return fmt.Sprintf("host=localhost port=15432 user=postgres password=%s %s sslmode=disable TimeZone=Asia/Tokyo", *pgpassword, dbnameAttr)
}

func connectDB(dbname string) (*pgx.Conn, error) {
	return pgx.Connect(context.Background(), formatDSN(dbname))
}

func createTestDatabase(dbname string) (string, error) {
	conn, err := connectDB("")
	if err != nil {
		return "", err
	}
	defer conn.Close(context.Background())

	if err := dropDatabaseIfExists(conn, dbname); err != nil {
		return "", err
	}

	if _, err := conn.Exec(context.Background(), "CREATE DATABASE "+dbname); err != nil {
		return "", err
	}

	return formatDSN(dbname), nil
}

func dropDatabaseIfExists(conn *pgx.Conn, dbname string) error {
	const SQL = "SELECT pid FROM pg_stat_activity WHERE datname = $1"
	rows, err := conn.Query(context.Background(), SQL, dbname)
	if err != nil {
		return err
	}
	defer rows.Close()

	pids := make([]int, 0)
	for rows.Next() {
		var pid int
		rows.Scan(&pid)
		pids = append(pids, pid)
	}
	rows.Close()

	for _, v := range pids {
		if _, err := conn.Exec(context.Background(), "SELECT pg_terminate_backend($1)", v); err != nil {
			return err
		}
	}

	if _, err := conn.Exec(context.Background(), "DROP DATABASE IF EXISTS "+dbname); err != nil {
		return err
	}

	return nil
}

func setupTest(testname string) (*echo.Echo, *sql.DB, pgx.Tx, []byte, int, error) {
	e, sqlDB, _, tx, jwtkey, validityMin, err := setupTestMain(testname, service.NewService)
	if err != nil {
		cleanDB(testname, sqlDB)
		sqlDB = nil
	}
	return e, sqlDB, tx, jwtkey, validityMin, err
}

func setupMockTest(testname string) (*echo.Echo, *sql.DB, *serviceMock, pgx.Tx, []byte, int, error) {
	newService := func(r repository.Repository) service.Service {
		return newMockService(newMockRepository(r))
	}
	e, sqlDB, s, tx, jwtkey, validityMin, err := setupTestMain(testname, newService)
	if err != nil {
		cleanDB(testname, sqlDB)
		sqlDB = nil
	}
	return e, sqlDB, s.(*serviceMock), tx, jwtkey, validityMin, err
}

func setupTestMain(testname string, newService func(repository.Repository) service.Service) (*echo.Echo, *sql.DB, service.Service, pgx.Tx, []byte, int, error) {
	// Database
	dbname := strings.ToLower(testname)
	dburl, err := createTestDatabase(dbname)
	if err != nil {
		return nil, nil, nil, nil, nil, 0, err
	}

	// Repository
	sqlDB, err := sql.Open("pgx", dburl)
	if err != nil {
		return nil, nil, nil, nil, nil, 0, err
	}
	r, err := repository.NewRepository(sqlDB)
	if err != nil {
		return nil, sqlDB, nil, nil, nil, 0, err
	}
	if err := r.InitDb(context.Background()); err != nil {
		return nil, sqlDB, nil, nil, nil, 0, err
	}

	// Service
	s := newService(r)

	// Handler
	jwtkey := []byte(testname)
	validityMin := 1
	location, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return nil, sqlDB, nil, nil, nil, 0, err
	}
	h := handler.NewHandler(s, &handler.HandlerConfig{
		JwtKey:      jwtkey,
		ValidityMin: validityMin,
		Location:    location,
		Indent:      "  ",
		TimeoutSec:  60,
	})

	// Echo
	e := handler.NewEcho(h)

	// トランザクション
	conn, err := connectDB(dbname)
	if err != nil {
		return nil, sqlDB, nil, nil, nil, 0, err
	}
	tx, err := conn.Begin(context.Background())
	if err != nil {
		return nil, sqlDB, nil, nil, nil, 0, err
	}

	return e, sqlDB, s, tx, jwtkey, validityMin, nil
}

func cleanIfSuccess(testname string, t *testing.T, sqlDB *sql.DB) error {
	if t.Failed() {
		sqlDB.Close()
		return nil
	}
	return cleanDB(testname, sqlDB)
}

func cleanDB(testname string, sqlDB *sql.DB) error {
	if sqlDB == nil {
		return nil
	}
	defer sqlDB.Close()

	conn, err := connectDB("")
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	dbname := strings.ToLower(testname)
	return dropDatabaseIfExists(conn, dbname)
}

func execHandlerTest(e *echo.Echo, tx pgx.Tx, req *http.Request) (*httptest.ResponseRecorder, *DiffTables, *TestDB, error) {
	loadedTime := time.Now()

	tables, err := loadTables(tx)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, nil, nil, err
	}

	if err := tx.Commit(context.Background()); err != nil {
		tx.Rollback(context.Background())
		return nil, nil, nil, err
	}

	before := &TestDB{
		loadedTime,
		tables,
	}

	rec, err := execHandler(e, req)
	if err != nil {
		return nil, nil, nil, err
	}

	conn := tx.Conn()
	defer conn.Close(context.Background())

	tx, err = conn.BeginTx(context.Background(), pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, nil, nil, err
	}
	defer tx.Rollback(context.Background())

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

func genToken(userId uint, jwtkey []byte, validityMin int) *string {
	iat := time.Now()
	claims := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&handler.JwtCustomClaims{
			jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(iat.Add(time.Duration(validityMin) * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(iat),
			},
			userId,
		},
	)

	signed, err := claims.SignedString(jwtkey)
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
	const SQL = "SELECT id, created_at, updated_at, deleted_at, user_id, date_time, store, product, price, in_stock FROM prices"
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
			&price.InStock,
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

func insertPrice(tx pgx.Tx, createdAt, updatedAt, deletedAt *time.Time, userID uint, dateTime time.Time, store, product string, price uint, inStock bool) (id uint, err error) {
	const SQL = "INSERT INTO prices (created_at, updated_at, deleted_at, user_id, date_time, store, product, price, in_stock) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id"
	err = tx.QueryRow(context.Background(), SQL, createdAt, updatedAt, deletedAt, userID, dateTime, store, product, price, inStock).Scan(&id)
	return
}

func insertPrices(tx pgx.Tx, t *time.Time, rows [][6]any) (int64, error) {
	inputRows := make([][]any, len(rows))
	for i, v := range rows {
		inputRows[i] = []any{t, t, v[0], v[1], v[2], v[3], v[4], v[5]}
	}
	return tx.CopyFrom(
		context.Background(),
		pgx.Identifier{"prices"},
		[]string{"created_at", "updated_at", "user_id", "date_time", "store", "product", "price", "in_stock"},
		pgx.CopyFromRows(inputRows),
	)
}
