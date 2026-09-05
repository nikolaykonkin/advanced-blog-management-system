package repository

import (
	"advanced-blog-management-system/internal/model"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type postRepository struct {
	db *sql.DB
}

// NewPostRepository создает новый репозиторий постов
func NewPostRepository(db *sql.DB) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(ctx context.Context, post *model.Post) error {
	now := time.Now().UTC()
	post.CreatedAt = now
	post.UpdatedAt = now

	query := `
		INSERT INTO posts (title, content, author_id, status, publish_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	err := r.db.QueryRowContext(ctx, query,
		post.Title, post.Content, post.AuthorID, post.Status, post.PublishAt, post.CreatedAt, post.UpdatedAt,
	).Scan(&post.ID)
	if err != nil {
		return fmt.Errorf("failed to create post: %w", err)
	}
	return nil
}

func (r *postRepository) GetByID(ctx context.Context, id int) (*model.Post, error) {
	query := `SELECT id, title, content, author_id, status, publish_at, created_at, updated_at FROM posts WHERE id = $1`

	var p model.Post
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Title, &p.Content, &p.AuthorID, &p.Status, &p.PublishAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get post by id: %w", err)
	}
	return &p, nil
}

func (r *postRepository) GetAll(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	query := `
		SELECT id, title, content, author_id, status, publish_at, created_at, updated_at
		FROM posts
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}
	defer rows.Close()

	posts := make([]*model.Post, 0)
	for rows.Next() {
		var p model.Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.AuthorID, &p.Status, &p.PublishAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate posts: %w", err)
	}
	return posts, nil
}

func (r *postRepository) GetTotalCount(ctx context.Context) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts`).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count posts: %w", err)
	}
	return count, nil
}

func (r *postRepository) Update(ctx context.Context, post *model.Post) error {
	post.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE posts
		SET title = $1, content = $2, status = $3, publish_at = $4, updated_at = $5
		WHERE id = $6
	`
	result, err := r.db.ExecContext(ctx, query, post.Title, post.Content, post.Status, post.PublishAt, post.UpdatedAt, post.ID)
	if err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("post with id %d not found", post.ID)
	}
	return nil
}

func (r *postRepository) Delete(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM posts WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("post with id %d not found", id)
	}
	return nil
}

func (r *postRepository) Exists(ctx context.Context, id int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)`
	if err := r.db.QueryRowContext(ctx, query, id).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check post existence: %w", err)
	}
	return exists, nil
}

func (r *postRepository) GetByAuthorID(ctx context.Context, authorID int, limit, offset int) ([]*model.Post, error) {
	query := `
		SELECT id, title, content, author_id, status, publish_at, created_at, updated_at
		FROM posts
		WHERE author_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, authorID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts by author: %w", err)
	}
	defer rows.Close()

	posts := make([]*model.Post, 0)
	for rows.Next() {
		var p model.Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.AuthorID, &p.Status, &p.PublishAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate posts: %w", err)
	}
	return posts, nil
}

func (r *postRepository) GetTotalCountByAuthorID(ctx context.Context, authorID int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM posts WHERE author_id = $1`
	if err := r.db.QueryRowContext(ctx, query, authorID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count posts by author: %w", err)
	}
	return count, nil
}

func (r *postRepository) GetScheduledPosts(ctx context.Context) ([]*model.Post, error) {
	query := `
		SELECT id, title, content, author_id, status, publish_at, created_at, updated_at
		FROM posts
		WHERE status = $1 AND publish_at IS NOT NULL AND publish_at <= NOW()
	`
	rows, err := r.db.QueryContext(ctx, query, model.PostStatusDraft)
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled posts: %w", err)
	}
	defer rows.Close()

	posts := make([]*model.Post, 0)
	for rows.Next() {
		var p model.Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.AuthorID, &p.Status, &p.PublishAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate scheduled posts: %w", err)
	}
	return posts, nil
}

func (r *postRepository) PublishPost(ctx context.Context, id int) error {
	query := `UPDATE posts SET status = $1, publish_at = NULL, updated_at = $2 WHERE id = $3`
	result, err := r.db.ExecContext(ctx, query, model.PostStatusPublished, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to publish post: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("post with id %d not found", id)
	}
	return nil
}
