package service

import (
	"context"

	"github.com/example/my-app-v2/ent"
)

type BookService struct{ client *ent.Client }

func NewBookService(client *ent.Client) *BookService {
	return &BookService{client: client}
}

func (s *BookService) Create(ctx context.Context, title, author string, price float64) (*ent.Book, error) {
	return s.client.Book.Create().
		SetTitle(title).
		SetNillableAuthor(&author).
		SetPrice(price).
		Save(ctx)
}

func (s *BookService) List(ctx context.Context) ([]*ent.Book, error) {
	return s.client.Book.Query().All(ctx)
}

func (s *BookService) GetByID(ctx context.Context, id int) (*ent.Book, error) {
	return s.client.Book.Get(ctx, id)
}

func (s *BookService) Update(ctx context.Context, id int, title, author string, price float64) (*ent.Book, error) {
	return s.client.Book.UpdateOneID(id).
		SetTitle(title).
		SetNillableAuthor(&author).
		SetPrice(price).
		Save(ctx)
}

func (s *BookService) Delete(ctx context.Context, id int) error {
	return s.client.Book.DeleteOneID(id).Exec(ctx)
}
