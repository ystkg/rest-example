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

	e, h, tx, _, _, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t)

	// mock
	mock := newMockRepository(h.Service().(*serviceMock))
	mock.user = newMockUserRepository(h.Service().(*serviceMock))
	mock.user.(*userRepositoryMock).err = errors.New(testname)
	h.SetService(newMockService(mock))

	now := time.Now()
	if _, err := insertUsers(tx, &now, someUsers()); err != nil {
		t.Fatal(err)
	}

	name, password := "testuser01", "testpassword"
	body := fmt.Sprintf("name=%s&password=%s", name, password)

	req := newRequest(
		http.MethodPost,
		"/user",
		&body,
		echo.MIMEApplicationForm,
		nil,
	)

	_, diff, _, err := execHandlerTest(e, tx, req)

	assert.Equal(t, mock.user.(*userRepositoryMock).err, err)
	assert.Nil(t, diff)
}

func TestCreateUserBeginTxError(t *testing.T) {
	testname := "TestCreateUserBeginTxError"

	e, h, tx, _, _, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t)

	// mock
	mock := newMockRepository(h.Service().(*serviceMock))
	mock.beginTxErr = errors.New(testname)
	h.SetService(newMockService(mock))

	now := time.Now()
	if _, err := insertUsers(tx, &now, someUsers()); err != nil {
		t.Fatal(err)
	}

	name, password := "testuser01", "testpassword"
	body := fmt.Sprintf("name=%s&password=%s", name, password)

	req := newRequest(
		http.MethodPost,
		"/user",
		&body,
		echo.MIMEApplicationForm,
		nil,
	)

	_, diff, _, err := execHandlerTest(e, tx, req)

	assert.Equal(t, mock.beginTxErr, err)
	assert.Nil(t, diff)
}

func TestCreateUserCommitError(t *testing.T) {
	testname := "TestCreateUserCommitError"

	e, h, tx, _, _, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t)

	// mock
	mock := newMockRepository(h.Service().(*serviceMock))
	mock.commitErr = errors.New(testname)
	h.SetService(newMockService(mock))

	now := time.Now()
	if _, err := insertUsers(tx, &now, someUsers()); err != nil {
		t.Fatal(err)
	}

	name, password := "testuser01", "testpassword"
	body := fmt.Sprintf("name=%s&password=%s", name, password)

	req := newRequest(
		http.MethodPost,
		"/user",
		&body,
		echo.MIMEApplicationForm,
		nil,
	)

	_, diff, _, err := execHandlerTest(e, tx, req)

	assert.Equal(t, mock.commitErr, err)
	assert.Nil(t, diff)
}

func TestGenTokenFindError(t *testing.T) {
	testname := "TestGenTokenFindError"

	e, h, tx, _, _, err := setupMockTest(testname)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanIfSuccess(testname, t)

	// mock
	mock := newMockRepository(h.Service().(*serviceMock))
	mock.user = newMockUserRepository(h.Service().(*serviceMock))
	mock.user.(*userRepositoryMock).err = errors.New(testname)
	h.SetService(newMockService(mock))

	now := time.Now()
	if _, err := insertUsers(tx, &now, someUsers()); err != nil {
		t.Fatal(err)
	}

	name, password := "testuser01", "testpassword"
	_, err = insertUser(tx, &now, &now, nil, name, encodePassword(name, password))
	if err != nil {
		t.Fatal(err)
	}

	body := "password=" + password

	req := newRequest(
		http.MethodPost,
		"/user/"+name+"/token",
		&body,
		echo.MIMEApplicationForm,
		nil,
	)

	_, diff, _, err := execHandlerTest(e, tx, req)

	assert.Equal(t, mock.user.(*userRepositoryMock).err, err)
	assert.Nil(t, diff)
}
