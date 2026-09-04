package repository

import (
	"advanced-blog-management-system/internal/model"
	"context"
	"database/sql"
)

type commentRepository struct {
	db *sql.DB
}

// NewCommentRepository создает новый репозиторий комментариев
func NewCommentRepository(db *sql.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(ctx context.Context, comment *model.Comment) error {
	// TODO: Реализовать создание комментария
	// Вставить комментарий в таблицу comments
	// Установить ID возвращаемое значение
	return nil
}

func (r *commentRepository) GetByID(ctx context.Context, id int) (*model.Comment, error) {
	// TODO: Реализовать получение комментария по ID
	return nil, nil
}

func (r *commentRepository) GetByPostID(ctx context.Context, postID int, limit, offset int) ([]*model.Comment, error) {
	// TODO: Реализовать получение комментариев к посту с пагинацией
	// Использовать LIMIT и OFFSET
	// Отсортировать по created_at ASC (старые сначала)
	return nil, nil
}

func (r *commentRepository) GetCountByPostID(ctx context.Context, postID int) (int, error) {
	// TODO: Реализовать получение количества комментариев к посту
	return 0, nil
}

func (r *commentRepository) Update(ctx context.Context, comment *model.Comment) error {
	// TODO: Реализовать обновление комментария
	// Обновить поле content
	// Обновить updated_at
	return nil
}

func (r *commentRepository) Delete(ctx context.Context, id int) error {
	// TODO: Реализовать удаление комментария
	return nil
}
