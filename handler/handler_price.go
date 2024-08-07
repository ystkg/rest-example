package handler

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/ystkg/rest-example/api"
	"github.com/ystkg/rest-example/entity"
	"github.com/ystkg/rest-example/service"

	"github.com/labstack/echo/v4"
)

// 価格の登録
func (h *Handler) CreatePrice(c echo.Context) error {
	// リクエストの取得
	userId := h.userId(c)
	req := &api.Price{}
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	inStock := true
	if req.InStock != nil {
		inStock = *req.InStock
	}

	// 入力チェック
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

	// サービスの実行
	price, err := h.service.CreatePrice(
		c.Request().Context(),
		userId,
		dateTime,
		req.Store,
		req.Product,
		req.Price,
		inStock,
	)
	if err != nil {
		return err
	}

	// レスポンスの生成
	return c.JSONPretty(http.StatusCreated, h.entityToResponse(price), h.indent)
}

// 価格の一覧
func (h *Handler) FindPrices(c echo.Context) error {
	// リクエストの取得
	userId := h.userId(c)

	// サービスの実行
	entities, err := h.service.FindPrices(c.Request().Context(), userId)
	if err != nil {
		return err
	}

	// レスポンスの生成
	sort.SliceStable(entities, func(i, j int) bool { return entities[i].DateTime.Unix() > entities[j].DateTime.Unix() }) // 降順
	priceList := make([]*api.Price, len(entities))
	for i, v := range entities {
		priceList[i] = h.entityToResponse(&v)
	}

	return c.JSONPretty(http.StatusOK, priceList, h.indent)
}

// 価格の取得
func (h *Handler) FindPrice(c echo.Context) error {
	// リクエストの取得
	userId := h.userId(c)
	reqId := c.Param("id")

	// 入力チェック
	priceId, err := strconv.ParseUint(reqId, 10, 0)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, ErrorNotFound.Error())
	}

	// サービスの実行
	price, err := h.service.FindPrice(c.Request().Context(), uint(priceId), userId)
	if err != nil {
		return err
	}
	if price == nil {
		return echo.NewHTTPError(http.StatusNotFound, ErrorNotFound.Error())
	}

	// レスポンスの生成
	return c.JSONPretty(http.StatusOK, h.entityToResponse(price), h.indent)
}

// 価格の更新
func (h *Handler) UpdatePrice(c echo.Context) error {
	// リクエストの取得
	userId := h.userId(c)
	reqId := c.Param("id")
	req := &api.Price{}
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	inStock := true
	if req.InStock != nil {
		inStock = *req.InStock
	}

	// 入力チェック
	priceId, err := strconv.ParseUint(reqId, 10, 0)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, ErrorNotFound.Error())
	}
	if err = c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.ID != nil && *req.ID != uint(priceId) {
		return echo.NewHTTPError(http.StatusBadRequest, ErrorIDUnchangeable.Error())
	}
	dateTime, err := h.parseDateTime(req.DateTime)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// サービスの実行
	price, err := h.service.UpdatePrice(
		c.Request().Context(),
		uint(priceId),
		userId,
		dateTime,
		req.Store,
		req.Product,
		req.Price,
		inStock,
	)
	if err != nil {
		if errors.Is(err, service.ErrorNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, ErrorNotFound.Error())
		}
		return err
	}

	// レスポンスの生成
	return c.JSONPretty(http.StatusOK, h.entityToResponse(price), h.indent)
}

// 価格の削除
func (h *Handler) DeletePrice(c echo.Context) error {
	// リクエストの取得
	userId := h.userId(c)
	reqId := c.Param("id")

	// 入力チェック
	priceId, err := strconv.ParseUint(reqId, 10, 0)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, ErrorNotFound.Error())
	}

	// サービスの実行
	if err = h.service.DeletePrice(c.Request().Context(), uint(priceId), userId); err != nil {
		if errors.Is(err, service.ErrorNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, ErrorNotFound.Error())
		}
		return err
	}

	// レスポンスの生成
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) parseDateTime(dateTime *string) (time.Time, error) {
	if dateTime == nil {
		return time.Now(), nil
	}
	return time.ParseInLocation(h.layout, *dateTime, h.location)
}

func (h *Handler) entityToResponse(entity *entity.Price) *api.Price {
	dateTime := entity.DateTime.Format(h.layout)
	return &api.Price{
		ID:       &entity.ID,
		DateTime: &dateTime,
		Store:    entity.Store,
		Product:  entity.Product,
		Price:    entity.Price,
		InStock:  &entity.InStock,
	}
}
