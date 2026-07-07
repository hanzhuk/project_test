package service

import (
	"context"
	"github.com/example/my-new-app/ent"
)

// BookService 提供图书业务逻辑操作。
type BookService struct {
	client *ent.Client
}

// NewBookService 创建新的 BookService 实例。
func NewBookService(client *ent.Client) *BookService {
	return &BookService{client: client}
}

// Create 创建新书目并返回创建的实体。
func (s *BookService) Create(ctx context.Context, title string, author string, isbn string) (*ent.Book, error) {
	book := s.client.Book.Create().SetTitle(title).SetAuthor(author).SetIsbn(isbn).Save(ctx)
	return book, nil
}

// List 查询所有图书列表。
func (s *BookService) List(ctx context.Context) ([]*ent.Book, error) {
	return s.client.Book.Query().All(ctx)
}

// GetByID 根据 ID 获取单本图书，不存在时返回错误。
func (s *BookService) GetByID(ctx context.Context, id int) (*ent.Book, error) {
	book, err := s.client.Book.Get(ctx, id)
	return book, err
}

// Update 更新指定 ID 的图书信息并保存结果。
func (s *BookService) Update(ctx context.Context, id int, title string, author string, isbn string) (*ent.Book, error) {
	book, err := s.client.Book.UpdateOneID(id).SetTitle(title).SetAuthor(author).SetIsbn(isbn).Save(ctx)
	return book, err
}

// Delete 删除指定 ID 的图书，不存在时返回错误。
func (s *BookService) Delete(ctx context.Context, id int) error {
	_, err := s.client.Book.DeleteOneID(id).Exec(ctx)
	return err
}