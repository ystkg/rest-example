package handler

import (
	"errors"
	"log/slog"
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
func (h *Handler) createPrice(c echo.Context) error {
	ctx := c.Request().Context()
	slog.DebugContext(ctx, "start")
	defer slog.DebugContext(ctx, "end")

	// リクエストの取得
	userId := h.userId(c)
	req := &api.Price{}
	if err := c.Bind(req); err != nil {
		return newHTTPError(http.StatusBadRequest, err)
	}

	// 入力チェック
	if err := c.Validate(req); err != nil {
		return newHTTPError(http.StatusBadRequest, err)
	}
	if req.ID != nil {
		return newHTTPError(http.StatusBadRequest, ErrIDCannotRequest)
	}
	dateTime, err := h.parseDateTime(req.DateTime)
	if err != nil {
		return newHTTPError(http.StatusBadRequest, err)
	}

	// サービスの実行
	price, err := h.service.CreatePrice(
		ctx,
		userId,
		dateTime,
		req.Store,
		req.Product,
		req.Price,
	)
	if err != nil {
		return err
	}

	// レスポンスの生成
	return c.JSONPretty(http.StatusCreated, h.entityToResponse(price), h.indent)
}

// 価格の一覧
func (h *Handler) findPrices(c echo.Context) error {
	ctx := c.Request().Context()
	slog.DebugContext(ctx, "start")
	defer slog.DebugContext(ctx, "end")

	// リクエストの取得
	userId := h.userId(c)

	// サービスの実行
	entities, err := h.service.FindPrices(ctx, userId)
	if err != nil {
		return err
	}

	// レスポンスの生成
	sort.SliceStable(entities, func(i, j int) bool {
		// 降順
		iDateTime := entities[i].DateTime.UnixNano()
		jDateTime := entities[j].DateTime.UnixNano()
		if iDateTime == jDateTime {
			return entities[i].ID > entities[j].ID // 第二ソートキー
		}
		return iDateTime > jDateTime // 第一ソートキー
	})
	priceList := make([]*api.Price, len(entities))
	for i, v := range entities {
		priceList[i] = h.entityToResponse(&v)
	}

	return c.JSONPretty(http.StatusOK, priceList, h.indent)
}

// 価格の取得
func (h *Handler) findPrice(c echo.Context) error {
	ctx := c.Request().Context()
	slog.DebugContext(ctx, "start")
	defer slog.DebugContext(ctx, "end")

	// リクエストの取得
	userId := h.userId(c)
	reqId := c.Param("id")

	// 入力チェック
	priceId, err := strconv.ParseUint(reqId, 10, 0)
	if err != nil {
		return newHTTPError(http.StatusNotFound, ErrNotFound)
	}

	// サービスの実行
	price, err := h.service.FindPrice(ctx, uint(priceId), userId)
	if err != nil {
		return err
	}
	if price == nil {
		return newHTTPError(http.StatusNotFound, ErrNotFound)
	}

	// レスポンスの生成
	return c.JSONPretty(http.StatusOK, h.entityToResponse(price), h.indent)
}

// 価格の更新
func (h *Handler) updatePrice(c echo.Context) error {
	ctx := c.Request().Context()
	slog.DebugContext(ctx, "start")
	defer slog.DebugContext(ctx, "end")

	// リクエストの取得
	userId := h.userId(c)
	reqId := c.Param("id")
	req := &api.Price{}
	if err := c.Bind(req); err != nil {
		return newHTTPError(http.StatusBadRequest, err)
	}

	// 入力チェック
	priceId, err := strconv.ParseUint(reqId, 10, 0)
	if err != nil {
		return newHTTPError(http.StatusNotFound, ErrNotFound)
	}
	if err = c.Validate(req); err != nil {
		return newHTTPError(http.StatusBadRequest, err)
	}
	if req.ID != nil && *req.ID != uint(priceId) {
		return newHTTPError(http.StatusBadRequest, ErrIDUnchangeable)
	}
	dateTime, err := h.parseDateTime(req.DateTime)
	if err != nil {
		return newHTTPError(http.StatusBadRequest, err)
	}

	// サービスの実行
	price, err := h.service.UpdatePrice(
		ctx,
		uint(priceId),
		userId,
		dateTime,
		req.Store,
		req.Product,
		req.Price,
	)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return newHTTPError(http.StatusNotFound, ErrNotFound)
		}
		return err
	}

	// レスポンスの生成
	return c.JSONPretty(http.StatusOK, h.entityToResponse(price), h.indent)
}

// 価格の削除
func (h *Handler) deletePrice(c echo.Context) error {
	ctx := c.Request().Context()
	slog.DebugContext(ctx, "start")
	defer slog.DebugContext(ctx, "end")

	// リクエストの取得
	userId := h.userId(c)
	reqId := c.Param("id")

	// 入力チェック
	priceId, err := strconv.ParseUint(reqId, 10, 0)
	if err != nil {
		return newHTTPError(http.StatusNotFound, ErrNotFound)
	}

	// サービスの実行
	if err = h.service.DeletePrice(ctx, uint(priceId), userId); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return newHTTPError(http.StatusNotFound, ErrNotFound)
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

func (h *Handler) formatDateTime(dateTime time.Time) string {
	return dateTime.In(h.location).Format(h.layout)
}

func (h *Handler) entityToResponse(entity *entity.Price) *api.Price {
	dateTime := h.formatDateTime(entity.DateTime)
	return &api.Price{
		ID:       &entity.ID,
		DateTime: &dateTime,
		Store:    entity.Store,
		Product:  entity.Product,
		Price:    entity.Price,
	}
}
