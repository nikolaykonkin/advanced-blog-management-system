package service

import (
	"advanced-blog-management-system/internal/errors/apperrors"
	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/repository"
	"context"
	"errors"
	"fmt"
	"time"
)

// commentDeletionPageSize — размер страницы при постраничном удалении комментариев поста
// var, не const: тесты временно уменьшают оба значения, чтобы не гонять цикл 100000 раз
var commentDeletionPageSize = 100

// maxCommentDeletionIterations — защитный предел итераций в deleteAllCommentsForPost
// на случай, если Delete вернет nil, ничего не удалив - иначе цикл станет бесконечным
var maxCommentDeletionIterations = 100000

type PostService struct {
	postRepo    repository.PostRepository
	userRepo    repository.UserRepository
	commentRepo repository.CommentRepository
	// actionLog — отложенный логгер действий пользователя
	// Может быть nil - вызовы в этом случае просто пропускаются
	actionLog ActionLogger
}

func NewPostService(postRepo repository.PostRepository, userRepo repository.UserRepository, commentRepo repository.CommentRepository, actionLog ActionLogger) *PostService {
	return &PostService{
		postRepo:    postRepo,
		userRepo:    userRepo,
		commentRepo: commentRepo,
		actionLog:   actionLog,
	}
}

// logAction отправляет событие в отложенный логгер, если он задан
func (s *PostService) logAction(event string) {
	if s.actionLog != nil {
		s.actionLog.Log(event)
	}
}

func (s *PostService) CreatePost(ctx context.Context, req *model.PostCreateRequest, authorID int) (*model.Post, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	post := &model.Post{
		Title:    req.Title,
		Content:  req.Content,
		AuthorID: authorID,
	}

	now := time.Now().UTC()
	if req.PublishAt == nil || !req.PublishAt.After(now) {
		post.Status = model.PostStatusPublished
	} else {
		publishAt := req.PublishAt.UTC()
		post.Status = model.PostStatusDraft
		post.PublishAt = &publishAt
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	s.logAction(fmt.Sprintf("user %d created post %d", authorID, post.ID))

	return post, nil
}

func (s *PostService) GetPost(ctx context.Context, id int) (*model.Post, error) {
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}
	if post == nil {
		return nil, apperrors.ErrPostNotFound
	}
	return post, nil
}

func (s *PostService) GetAllPosts(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	posts, err := s.postRepo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}
	return posts, nil
}

func (s *PostService) GetPostsCount(ctx context.Context) (int, error) {
	count, err := s.postRepo.GetTotalCount(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count posts: %w", err)
	}
	return count, nil
}

func (s *PostService) UpdatePost(ctx context.Context, id int, req *model.PostUpdateRequest, userID int) (*model.Post, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}
	if post == nil {
		return nil, apperrors.ErrPostNotFound
	}
	if !post.CanBeEditedBy(userID) {
		return nil, apperrors.ErrForbidden
	}

	post.Title = req.Title
	post.Content = req.Content
	if req.Status != "" {
		post.Status = req.Status
	}
	if req.PublishAt != nil {
		publishAt := req.PublishAt.UTC()
		post.PublishAt = &publishAt
	} else {
		post.PublishAt = nil
	}

	if err := s.postRepo.Update(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	return post, nil
}

func (s *PostService) DeletePost(ctx context.Context, id int, userID int) error {
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get post: %w", err)
	}
	if post == nil {
		return apperrors.ErrPostNotFound
	}
	if !post.CanBeDeletedBy(userID) {
		return apperrors.ErrForbidden
	}

	if err := s.deleteAllCommentsForPost(ctx, id); err != nil {
		return fmt.Errorf("failed to delete comments for post: %w", err)
	}

	if err := s.postRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	return nil
}

// deleteAllCommentsForPost вычищает все комментарии к посту постранично
// offset намеренно всегда 0: после удаления страницы следующий запрос с тем же
// offset=0 видит уже новую "первую страницу" оставшихся комментариев, а не пропускает их
func (s *PostService) deleteAllCommentsForPost(ctx context.Context, postID int) error {
	for i := 0; i < maxCommentDeletionIterations; i++ {
		comments, err := s.commentRepo.GetByPostID(ctx, postID, commentDeletionPageSize, 0)
		if err != nil {
			return err
		}
		if len(comments) == 0 {
			return nil
		}
		for _, c := range comments {
			if err := s.commentRepo.Delete(ctx, c.ID); err != nil {
				return err
			}
		}
	}
	return fmt.Errorf("exceeded %d iterations while deleting comments for post %d — Delete may not be removing rows", maxCommentDeletionIterations, postID)
}

func (s *PostService) GetPostsByAuthor(ctx context.Context, authorID int, limit, offset int) ([]*model.Post, error) {
	posts, err := s.postRepo.GetByAuthorID(ctx, authorID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts by author: %w", err)
	}
	return posts, nil
}

func (s *PostService) GetPostsCountByAuthor(ctx context.Context, authorID int) (int, error) {
	count, err := s.postRepo.GetTotalCountByAuthorID(ctx, authorID)
	if err != nil {
		return 0, fmt.Errorf("failed to count posts by author: %w", err)
	}
	return count, nil
}

// PublishScheduledPosts публикует все черновики, время публикации которых уже наступило
// Ошибка одного поста не прерывает цикл - остальные готовые посты все равно публикуются,
// иначе один битый пост блокировал бы всю очередь на каждом тике планировщика
func (s *PostService) PublishScheduledPosts(ctx context.Context) (int, error) {
	posts, err := s.postRepo.GetScheduledPosts(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get scheduled posts: %w", err)
	}

	published := 0
	var errs []error
	for _, post := range posts {
		// защита от пограничного случая, когда publish_at совпадает с моментом NOW() в SQL-запросе
		if !post.ShouldPublishNow() {
			continue
		}
		if err := s.postRepo.PublishPost(ctx, post.ID); err != nil {
			errs = append(errs, fmt.Errorf("failed to publish post %d: %w", post.ID, err))
			continue
		}
		published++
	}

	// errors.Join возвращает nil, если errs пуст, так что успешный путь не меняется
	return published, errors.Join(errs...)
}
