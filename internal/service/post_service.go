package service

import (
	"advanced-blog-management-system/internal/errors/apperrors"
	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/repository"
	"context"
	"fmt"
	"time"
)

const commentDeletionPageSize = 100

type PostService struct {
	postRepo    repository.PostRepository
	userRepo    repository.UserRepository
	commentRepo repository.CommentRepository
}

func NewPostService(postRepo repository.PostRepository, userRepo repository.UserRepository, commentRepo repository.CommentRepository) *PostService {
	return &PostService{
		postRepo:    postRepo,
		userRepo:    userRepo,
		commentRepo: commentRepo,
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

// deleteAllCommentsForPost вычищает все комментарии к посту постранично.
// offset намеренно всегда 0: после удаления очередной страницы эти строки
// исчезают из таблицы, и следующий запрос с тем же offset=0 забирает уже
// новую "первую страницу" оставшихся комментариев — а не пропускает их,
// как было бы при обычной постраничной навигации по неизменным данным.
func (s *PostService) deleteAllCommentsForPost(ctx context.Context, postID int) error {
	for {
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
// Не часть исходного TODO этого файла — добавлено для фонового планировщика
// (см. runScheduler в api/main.go), который периодически вызывает этот метод
func (s *PostService) PublishScheduledPosts(ctx context.Context) (int, error) {
	posts, err := s.postRepo.GetScheduledPosts(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get scheduled posts: %w", err)
	}

	published := 0
	for _, post := range posts {
		// Двойная проверка поверх SQL-фильтра GetScheduledPosts: используем
		// Post.ShouldPublishNow()из internal/model — защита от пограничного случая,
		// когда publish_at совпадает с моментом выполнения SQL-запроса NOW(),
		// и даёт этому методу модели реальное применение в коде, а не только в тестах
		if !post.ShouldPublishNow() {
			continue
		}
		if err := s.postRepo.PublishPost(ctx, post.ID); err != nil {
			return published, fmt.Errorf("failed to publish post %d: %w", post.ID, err)
		}
		published++
	}

	return published, nil
}
