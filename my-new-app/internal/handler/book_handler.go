package handler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/example/my-new-app/internal/response"
)

type CreateBookInput struct {
	Body struct {
		Title  string  `json:"title"`
		Author string  `json:"author"`
		Price  float64 `json:"price"`
	}
}

type BookResponse struct {
	Body response.Response
}

type ListBooksInput struct{}

type BookListResponse struct {
	Body response.Response
}

type GetBookInput struct {
	ID int `path:"id"`
}

type UpdateBookInput struct {
	ID   int `path:"id"`
	Body struct {
		Title  string  `json:"title"`
		Author string  `json:"author"`
		Price  float64 `json:"price"`
	}
}

type DeleteBookInput struct {
	ID int `path:"id"`
}

func (h *Handler) CreateBook(ctx context.Context, input *CreateBookInput) (*BookResponse, error) {
	book, err := h.Client.Book.Create().
		SetTitle(input.Body.Title).
		SetAuthor(input.Body.Author).
		SetPrice(input.Body.Price).
		Save(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "create book failed", "err", err)
		return nil, huma.Error500InternalServerError("创建失败", err)
	}
	return &BookResponse{Body: response.Success(book)}, nil
}

func (h *Handler) ListBooks(ctx context.Context, input *ListBooksInput) (*BookListResponse, error) {
	books, err := h.Client.Book.Query().All(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "list books failed", "err", err)
		return nil, huma.Error500InternalServerError("查询失败", err)
	}
	return &BookListResponse{Body: response.Success(books)}, nil
}

func (h *Handler) GetBook(ctx context.Context, input *GetBookInput) (*BookResponse, error) {
	book, err := h.Client.Book.Get(ctx, input.ID)
	if err != nil {
		return nil, huma.Error404NotFound("书籍不存在", fmt.Errorf("id=%d", input.ID))
	}
	return &BookResponse{Body: response.Success(book)}, nil
}

func (h *Handler) UpdateBook(ctx context.Context, input *UpdateBookInput) (*BookResponse, error) {
	book, err := h.Client.Book.UpdateOneID(input.ID).
		SetTitle(input.Body.Title).
		SetAuthor(input.Body.Author).
		SetPrice(input.Body.Price).
		Save(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "update book failed", "err", err)
		return nil, huma.Error500InternalServerError("更新失败", err)
	}
	return &BookResponse{Body: response.Success(book)}, nil
}

func (h *Handler) DeleteBook(ctx context.Context, input *DeleteBookInput) (*BookResponse, error) {
	if err := h.Client.Book.DeleteOneID(input.ID).Exec(ctx); err != nil {
		return nil, huma.Error404NotFound("书籍不存在", fmt.Errorf("id=%d", input.ID))
	}
	return &BookResponse{Body: response.Success(nil)}, nil
}
