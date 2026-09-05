package repository

import (
	"advanced-blog-management-system/internal/model"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type commentRepository struct {
	db *sql.DB
}

// NewCommentRepository создает новый репозиторий комментариев
func NewCommentRepository(db *sql.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(ctx context.Context, comment *model.Comment) error {
	now := time.Now().UTC()
	comment.CreatedAt = now
	comment.UpdatedAt = now

	query := `
		INSERT INTO comments (content, post_id, author_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	err := r.db.QueryRowContext(ctx, query,
		comment.Content, comment.PostID, comment.AuthorID, comment.CreatedAt, comment.UpdatedAt,
	).Scan(&comment.ID)
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}
	return nil
}

func (r *commentRepository) GetByID(ctx context.Context, id int) (*model.Comment, error) {
	query := `SELECT id, content, post_id, author_id, created_at, updated_at FROM comments WHERE id = $1`

	var c model.Comment
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.Content, &c.PostID, &c.AuthorID, &c.CreatedAt, &c.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get comment by id: %w", err)
	}
	return &c, nil
}

func (r *commentRepository) GetByPostID(ctx context.Context, postID int, limit, offset int) ([]*model.Comment, error) {
	query := `
		SELECT id, content, post_id, author_id, created_at, updated_at
		FROM comments
		WHERE post_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, postID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments by post: %w", err)
	}
	defer rows.Close()

	comments := make([]*model.Comment, 0)
	for rows.Next() {
		var c model.Comment
		if err := rows.Scan(&c.ID, &c.Content, &c.PostID, &c.AuthorID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate comments: %w", err)
	}
	return comments, nil
}

func (r *commentRepository) GetCountByPostID(ctx context.Context, postID int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM comments WHERE post_id = $1`
	if err := r.db.QueryRowContext(ctx, query, postID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count comments: %w", err)
	}
	return count, nil
}

func (r *commentRepository) Update(ctx context.Context, comment *model.Comment) error {
	comment.UpdatedAt = time.Now().UTC()

	query := `UPDATE comments SET content = $1, updated_at = $2 WHERE id = $3`
	result, err := r.db.ExecContext(ctx, query, comment.Content, comment.UpdatedAt, comment.ID)
	if err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("comment with id %d not found", comment.ID)
	}
	return nil
}

func (r *commentRepository) Delete(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM comments WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("comment with id %d not found", id)
	}
	return nil
}
