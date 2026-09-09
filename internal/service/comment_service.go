package service

import (
	"advanced-blog-management-system/internal/errors/apperrors"
	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/repository"
	"context"
	"errors"
	"fmt"
)

// ErrPostNotPublished возвращается при попытке оставить комментарий
// к посту, который ещё не опубликован (черновик/отложенная публикация)
var ErrPostNotPublished = errors.New("post is not published yet")

type CommentService struct {
	commentRepo repository.CommentRepository
	postRepo    repository.PostRepository
	userRepo    repository.UserRepository
	// actionLog — отложенный логгер действий пользователя
	// Может быть nil — тогда вызовы просто пропускаются
	actionLog ActionLogger
}

func NewCommentService(commentRepo repository.CommentRepository, postRepo repository.PostRepository, userRepo repository.UserRepository, actionLog ActionLogger) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		postRepo:    postRepo,
		userRepo:    userRepo,
		actionLog:   actionLog,
	}
}

// logAction отправляет событие в отложенный логгер, если он задан
func (s *CommentService) logAction(event string) {
	if s.actionLog != nil {
		s.actionLog.Log(event)
	}
}

func (s *CommentService) CreateComment(ctx context.Context, req *model.CommentCreateRequest, postID int, authorID int) (*model.Comment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}
	if post == nil {
		return nil, apperrors.ErrPostNotFound
	}
	if post.Status != model.PostStatusPublished {
		return nil, ErrPostNotPublished
	}

	comment := &model.Comment{
		Content:  req.Content,
		PostID:   postID,
		AuthorID: authorID,
	}
	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	s.logAction(fmt.Sprintf("user %d created comment %d", authorID, comment.ID))

	return comment, nil
}

func (s *CommentService) GetComment(ctx context.Context, id int) (*model.Comment, error) {
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}
	if comment == nil {
		return nil, apperrors.ErrCommentNotFound
	}
	return comment, nil
}

func (s *CommentService) GetCommentsByPostID(ctx context.Context, postID int, limit, offset int) ([]*model.Comment, error) {
	exists, err := s.postRepo.Exists(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to check post existence: %w", err)
	}
	if !exists {
		return nil, apperrors.ErrPostNotFound
	}

	comments, err := s.commentRepo.GetByPostID(ctx, postID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}
	return comments, nil
}

func (s *CommentService) GetCommentsCountByPostID(ctx context.Context, postID int) (int, error) {
	count, err := s.commentRepo.GetCountByPostID(ctx, postID)
	if err != nil {
		return 0, fmt.Errorf("failed to count comments: %w", err)
	}
	return count, nil
}

func (s *CommentService) UpdateComment(ctx context.Context, id int, req *model.CommentUpdateRequest, userID int) (*model.Comment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}
	if comment == nil {
		return nil, apperrors.ErrCommentNotFound
	}
	if !comment.CanBeEditedBy(userID) {
		return nil, apperrors.ErrForbidden
	}

	comment.Content = req.Content
	if err := s.commentRepo.Update(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	return comment, nil
}

func (s *CommentService) DeleteComment(ctx context.Context, id int, userID int) error {
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get comment: %w", err)
	}
	if comment == nil {
		return apperrors.ErrCommentNotFound
	}
	if !comment.CanBeDeletedBy(userID) {
		return apperrors.ErrForbidden
	}

	if err := s.commentRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	return nil
}
