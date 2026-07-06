package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/example/my-app/internal/response"
	"github.com/labstack/echo/v4"
)

type bookRequest struct {
	Title  string  `json:"title"`
	Author string  `json:"author"`
	Price  float64 `json:"price"`
}

func (h *Handler) CreateBook(c echo.Context) error {
	var req bookRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(400, "参数解析失败"))
	}
	book, err := h.Client.Book.Create().
		SetTitle(req.Title).
		SetAuthor(req.Author).
		SetPrice(req.Price).
		Save(c.Request().Context())
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "create book failed", "err", err)
		return c.JSON(http.StatusInternalServerError, response.Error(500, "创建失败"))
	}
	return c.JSON(http.StatusCreated, response.Success(book))
}

func (h *Handler) ListBooks(c echo.Context) error {
	books, err := h.Client.Book.Query().All(c.Request().Context())
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "list books failed", "err", err)
		return c.JSON(http.StatusInternalServerError, response.Error(500, "查询失败"))
	}
	return c.JSON(http.StatusOK, response.Success(books))
}

func (h *Handler) GetBook(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(400, "无效的 ID"))
	}
	book, err := h.Client.Book.Get(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, response.Error(404, "书籍不存在"))
	}
	return c.JSON(http.StatusOK, response.Success(book))
}

func (h *Handler) UpdateBook(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(400, "无效的 ID"))
	}
	var req bookRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(400, "参数解析失败"))
	}
	book, err := h.Client.Book.UpdateOneID(id).
		SetTitle(req.Title).
		SetAuthor(req.Author).
		SetPrice(req.Price).
		Save(c.Request().Context())
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "update book failed", "err", err)
		return c.JSON(http.StatusInternalServerError, response.Error(500, "更新失败"))
	}
	return c.JSON(http.StatusOK, response.Success(book))
}

func (h *Handler) DeleteBook(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(400, "无效的 ID"))
	}
	if err := h.Client.Book.DeleteOneID(id).Exec(c.Request().Context()); err != nil {
		slog.ErrorContext(c.Request().Context(), "delete book failed", "err", err)
		return c.JSON(http.StatusInternalServerError, response.Error(500, "删除失败"))
	}
	return c.JSON(http.StatusOK, response.Success(nil))
}
