package handler

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/ystkg/rest-example/api"
	"github.com/ystkg/rest-example/entity"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func (h *Handler) CreatePrice(c echo.Context) error {
	claims := c.Get("user").(*jwt.Token).Claims.(*JwtCustomClaims)
	req := &api.Price{}
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.ID != nil {
		return echo.NewHTTPError(http.StatusBadRequest, ErrorIDCannotRequest.Error())
	}
	dateTime, err := h.parseDateTime(req.DateTime)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	inStock := true
	if req.InStock != nil {
		inStock = *req.InStock
	}

	price := &entity.Price{
		UserID:   claims.UserId,
		DateTime: dateTime,
		Store:    req.Store,
		Product:  req.Product,
		Price:    req.Price,
		InStock:  inStock,
	}

	tx, err := h.beginTX()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if tx = tx.Create(price); tx.Error != nil {
		err := tx.Error
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return c.JSONPretty(http.StatusCreated, h.entityToResponse(price), "  ")
}

func (h *Handler) FindPrices(c echo.Context) error {
	claims := c.Get("user").(*jwt.Token).Claims.(*JwtCustomClaims)

	var entities []entity.Price
	db := h.db.Where("user_id = ?", claims.UserId).Find(&entities)
	if db.Error != nil {
		return db.Error
	}

	priceList := make([]*api.Price, db.RowsAffected)
	for i, v := range entities {
		priceList[i] = h.entityToResponse(&v)
	}
	sort.SliceStable(priceList, func(i, j int) bool { return entities[j].DateTime.Unix() < entities[i].DateTime.Unix() }) // desc

	return c.JSONPretty(http.StatusOK, priceList, "  ")
}

func (h *Handler) FindPrice(c echo.Context) error {
	claims := c.Get("user").(*jwt.Token).Claims.(*JwtCustomClaims)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, ErrorNotFound.Error())
	}

	price := &entity.Price{
		Model: gorm.Model{
			ID: uint(id),
		},
	}

	if db := h.db.Where("user_id = ?", claims.UserId).First(price); db.Error != nil {
		if errors.Is(db.Error, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, ErrorNotFound.Error())
		}
		return db.Error
	}

	return c.JSONPretty(http.StatusOK, h.entityToResponse(price), "  ")
}

func (h *Handler) UpdatePrice(c echo.Context) error {
	claims := c.Get("user").(*jwt.Token).Claims.(*JwtCustomClaims)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, ErrorNotFound.Error())
	}
	req := &api.Price{}
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.ID != nil && *req.ID != uint(id) {
		return echo.NewHTTPError(http.StatusBadRequest, ErrorIDUnchangeable.Error())
	}
	dateTime, err := h.parseDateTime(req.DateTime)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	inStock := true
	if req.InStock != nil {
		inStock = *req.InStock
	}

	price := &entity.Price{
		Model: gorm.Model{
			ID: uint(id),
		},
		UserID:   claims.UserId,
		DateTime: dateTime,
		Store:    req.Store,
		Product:  req.Product,
		Price:    req.Price,
		InStock:  inStock,
	}

	tx, err := h.beginTX()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if tx = tx.Where("user_id = ? and deleted_at is null", claims.UserId).Updates(price); tx.Error != nil {
		tx.Rollback()
		return tx.Error
	}
	rows := tx.RowsAffected
	if rows != 1 {
		tx.Rollback()
		if rows == 0 {
			return echo.NewHTTPError(http.StatusNotFound, ErrorNotFound.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("RowsAffected:%d", tx.RowsAffected))
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return c.JSONPretty(http.StatusOK, h.entityToResponse(price), "  ")
}

func (h *Handler) DeletePrice(c echo.Context) error {
	claims := c.Get("user").(*jwt.Token).Claims.(*JwtCustomClaims)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, ErrorNotFound.Error())
	}

	price := &entity.Price{
		Model: gorm.Model{
			ID: uint(id),
		},
	}

	tx, err := h.beginTX()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if tx = tx.Where("user_id = ?", claims.UserId).Delete(price); tx.Error != nil {
		tx.Rollback()
		return tx.Error
	}
	rows := tx.RowsAffected
	if rows != 1 {
		tx.Rollback()
		if rows == 0 {
			return echo.NewHTTPError(http.StatusNotFound, ErrorNotFound.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("RowsAffected:%d", tx.RowsAffected))
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) parseDateTime(dateTime *string) (time.Time, error) {
	if dateTime == nil {
		return time.Now(), nil
	}
	return time.ParseInLocation(h.layout, *dateTime, h.location)
}

func (h *Handler) entityToResponse(entity *entity.Price) *api.Price {
	id := entity.ID
	dateTime := entity.DateTime.Format(h.layout)
	inStock := entity.InStock
	return &api.Price{
		ID:       &id,
		DateTime: &dateTime,
		Store:    entity.Store,
		Product:  entity.Product,
		Price:    entity.Price,
		InStock:  &inStock,
	}
}
