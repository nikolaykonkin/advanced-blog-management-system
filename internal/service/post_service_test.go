package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"advanced-blog-management-system/internal/errors/apperrors"
	"advanced-blog-management-system/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestPostService() (*PostService, *fakePostRepository, *fakeCommentRepository, *fakeActionLogger) {
	postRepo := newFakePostRepository()
	userRepo := newFakeUserRepository()
	commentRepo := newFakeCommentRepository()
	logger := &fakeActionLogger{}
	svc := NewPostService(postRepo, userRepo, commentRepo, logger)
	return svc, postRepo, commentRepo, logger
}

func TestPostService_CreatePost_NoPublishAt_StatusPublished(t *testing.T) {
	svc, _, _, _ := newTestPostService()

	post, err := svc.CreatePost(context.Background(), &model.PostCreateRequest{Title: "Title", Content: "Content"}, 1)

	require.NoError(t, err)
	assert.Equal(t, model.PostStatusPublished, post.Status)
	assert.Nil(t, post.PublishAt)
}

func TestPostService_CreatePost_PublishAtInFuture_StatusDraft(t *testing.T) {
	svc, _, _, _ := newTestPostService()
	future := time.Now().Add(1 * time.Hour)

	post, err := svc.CreatePost(context.Background(), &model.PostCreateRequest{
		Title: "Title", Content: "Content", PublishAt: &future,
	}, 1)

	require.NoError(t, err)
	assert.Equal(t, model.PostStatusDraft, post.Status)
	require.NotNil(t, post.PublishAt)
}

func TestPostService_CreatePost_PublishAtInPast_StatusPublished(t *testing.T) {
	svc, _, _, _ := newTestPostService()
	past := time.Now().Add(-1 * time.Hour)

	post, err := svc.CreatePost(context.Background(), &model.PostCreateRequest{
		Title: "Title", Content: "Content", PublishAt: &past,
	}, 1)

	require.NoError(t, err)
	assert.Equal(t, model.PostStatusPublished, post.Status)
}

func TestPostService_CreatePost_InvalidRequest_ReturnsValidationError(t *testing.T) {
	svc, _, _, _ := newTestPostService()

	_, err := svc.CreatePost(context.Background(), &model.PostCreateRequest{Title: "", Content: "Content"}, 1)

	assert.Error(t, err)
}

func TestPostService_CreatePost_LogsActionEvent(t *testing.T) {
	svc, _, _, logger := newTestPostService()

	post, err := svc.CreatePost(context.Background(), &model.PostCreateRequest{Title: "Title", Content: "Content"}, 7)
	require.NoError(t, err)

	require.Len(t, logger.events, 1)
	assert.Equal(t, fmt.Sprintf("user 7 created post %d", post.ID), logger.events[0])
}

func TestPostService_GetPost_NotFound_ReturnsErrPostNotFound(t *testing.T) {
	svc, _, _, _ := newTestPostService()

	_, err := svc.GetPost(context.Background(), 999)

	assert.ErrorIs(t, err, apperrors.ErrPostNotFound)
}

func TestPostService_CreatePost_RepoError_ReturnsError(t *testing.T) {
	svc, postRepo, _, _ := newTestPostService()
	postRepo.createErr = errors.New("disk full")

	_, err := svc.CreatePost(context.Background(), &model.PostCreateRequest{Title: "Title", Content: "Content"}, 1)

	assert.Error(t, err)
}

func TestPostService_GetPost_RepoError_ReturnsError(t *testing.T) {
	svc, postRepo, _, _ := newTestPostService()
	postRepo.getByIDErr = errors.New("db unavailable")

	_, err := svc.GetPost(context.Background(), 1)

	assert.Error(t, err)
}

func TestPostService_GetAllPosts_Success(t *testing.T) {
	svc, postRepo, _, _ := newTestPostService()
	postRepo.posts[1] = &model.Post{ID: 1, Title: "One"}
	postRepo.posts[2] = &model.Post{ID: 2, Title: "Two"}

	posts, err := svc.GetAllPosts(context.Background(), 10, 0)

	require.NoError(t, err)
	assert.Len(t, posts, 2)
}

func TestPostService_GetAllPosts_RepoError_ReturnsError(t *testing.T) {
	svc, postRepo, _, _ := newTestPostService()
	postRepo.getAllErr = errors.New("db unavailable")

	_, err := svc.GetAllPosts(context.Background(), 10, 0)

	assert.Error(t, err)
}

func TestPostService_GetPostsCount_Success(t *testing.T) {
	svc, postRepo, _, _ := newTestPostService()
	postRepo.posts[1] = &model.Post{ID: 1}
	postRepo.posts[2] = &model.Post{ID: 2}

	count, err := svc.GetPostsCount(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestPostService_GetPostsCount_RepoError_ReturnsError(t *testing.T) {
	svc, postRepo, _, _ := newTestPostService()
	postRepo.getTotalCountErr = errors.New("db unavailable")

	_, err := svc.GetPostsCount(context.Background())

	assert.Error(t, err)
}

func TestPostService_GetPostsByAuthor_Success(t *testing.T) {
	svc, postRepo, _, _ := newTestPostService()
	postRepo.posts[1] = &model.Post{ID: 1, AuthorID: 5}
	postRepo.posts[2] = &model.Post{ID: 2, AuthorID: 9}

	posts, err := svc.GetPostsByAuthor(context.Background(), 5, 10, 0)

	require.NoError(t, err)
	require.Len(t, posts, 1)
	assert.Equal(t, 5, posts[0].AuthorID)
}

func TestPostService_GetPostsByAuthor_RepoError_ReturnsError(t *testing.T) {
	svc, postRepo, _, _ := newTestPostService()
	postRepo.getByAuthorIDErr = errors.New("db unavailable")

	_, err := svc.GetPostsByAuthor(context.Background(), 5, 10, 0)

	assert.Error(t, err)
}

func TestPostService_GetPostsCountByAuthor_Success(t *testing.T) {
	svc, postRepo, _, _ := newTestPostService()
	postRepo.posts[1] = &model.Post{ID: 1, AuthorID: 5}

	count, err := svc.GetPostsCountByAuthor(context.Background(), 5)

	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestPostService_GetPostsCountByAuthor_RepoError_ReturnsError(t *testing.T) {
	svc, postRepo, _, _ := newTestPostService()
	postRepo.getTotalCountByAuthorIDErr = errors.New("db unavailable")

	_, err := svc.GetPostsCountByAuthor(context.Background(), 5)

	assert.Error(t, err)
}

func TestPostService_UpdatePost_NotOwner_ReturnsErrForbidden(t *testing.T) {
	svc, postRepo, _, _ := newTestPostService()
	postRepo.posts[1] = &model.Post{ID: 1, AuthorID: 42, Title: "Old", Content: "Old"}

	_, err := svc.UpdatePost(context.Background(), 1, &model.PostUpdateRequest{Title: "New", Content: "New"}, 1)

	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestPostService_UpdatePost_NotFound_ReturnsErrPostNotFound(t *testing.T) {
	svc, _, _, _ := newTestPostService()

	_, err := svc.UpdatePost(context.Background(), 999, &model.PostUpdateRequest{Title: "New", Content: "New"}, 1)

	assert.ErrorIs(t, err, apperrors.ErrPostNotFound)
}

func TestPostService_UpdatePost_Owner_UpdatesFields(t *testing.T) {
	svc, postRepo, _, _ := newTestPostService()
	postRepo.posts[1] = &model.Post{ID: 1, AuthorID: 1, Title: "Old", Content: "Old"}

	post, err := svc.UpdatePost(context.Background(), 1, &model.PostUpdateRequest{Title: "New", Content: "New"}, 1)

	require.NoError(t, err)
	assert.Equal(t, "New", post.Title)
	assert.Equal(t, "New", post.Content)
}

func TestPostService_UpdatePost_InvalidRequest_ReturnsValidationError(t *testing.T) {
	svc, postRepo, _, _ := newTestPostService()
	postRepo.posts[1] = &model.Post{ID: 1, AuthorID: 1, Title: "Old", Content: "Old"}

	_, err := svc.UpdatePost(context.Background(), 1, &model.PostUpdateRequest{Title: "", Content: "New"}, 1)

	assert.Error(t, err)
}

func TestPostService_UpdatePost_RepoGetByIDError_ReturnsError(t *testing.T) {
	svc, postRepo, _, _ := newTestPostService()
	postRepo.getByIDErr = errors.New("db unavailable")

	_, err := svc.UpdatePost(context.Background(), 1, &model.PostUpdateRequest{Title: "New", Content: "New"}, 1)

	assert.Error(t, err)
}

func TestPostService_UpdatePost_RepoUpdateError_ReturnsError(t *testing.T) {
	svc, postRepo, _, _ := newTestPostService()
	postRepo.posts[1] = &model.Post{ID: 1, AuthorID: 1, Title: "Old", Content: "Old"}
	postRepo.updateErr = errors.New("disk full")

	_, err := svc.UpdatePost(context.Background(), 1, &model.PostUpdateRequest{Title: "New", Content: "New"}, 1)

	assert.Error(t, err)
}

func TestPostService_DeletePost_NotOwner_ReturnsErrForbidden(t *testing.T) {
	svc, postRepo, _, _ := newTestPostService()
	postRepo.posts[1] = &model.Post{ID: 1, AuthorID: 42}

	err := svc.DeletePost(context.Background(), 1, 1)

	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestPostService_DeletePost_NotFound_ReturnsErrPostNotFound(t *testing.T) {
	svc, _, _, _ := newTestPostService()

	err := svc.DeletePost(context.Background(), 999, 1)

	assert.ErrorIs(t, err, apperrors.ErrPostNotFound)
}

func TestPostService_DeletePost_Owner_DeletesPostAndItsComments(t *testing.T) {
	svc, postRepo, commentRepo, _ := newTestPostService()
	postRepo.posts[1] = &model.Post{ID: 1, AuthorID: 1}
	commentRepo.comments[1] = &model.Comment{ID: 1, PostID: 1}
	commentRepo.comments[2] = &model.Comment{ID: 2, PostID: 1}
	commentRepo.comments[3] = &model.Comment{ID: 3, PostID: 2} // другой пост - не должен удалиться

	err := svc.DeletePost(context.Background(), 1, 1)
	require.NoError(t, err)

	assert.NotContains(t, postRepo.posts, 1)
	assert.NotContains(t, commentRepo.comments, 1)
	assert.NotContains(t, commentRepo.comments, 2)
	assert.Contains(t, commentRepo.comments, 3)
}

func TestPostService_DeletePost_RepoGetByIDError_ReturnsError(t *testing.T) {
	svc, postRepo, _, _ := newTestPostService()
	postRepo.getByIDErr = errors.New("db unavailable")

	err := svc.DeletePost(context.Background(), 1, 1)

	assert.Error(t, err)
}

// В отличие от TestPostService_DeletePost_CommentDeletionExceedsIterationLimit
// (Delete сообщает об успехе, но ничего не удаляет),
// здесь Delete возвращает настоящую ошибку сразу
func TestPostService_DeletePost_CommentDeleteError_ReturnsError(t *testing.T) {
	svc, postRepo, commentRepo, _ := newTestPostService()
	postRepo.posts[1] = &model.Post{ID: 1, AuthorID: 1}
	commentRepo.comments[1] = &model.Comment{ID: 1, PostID: 1}
	commentRepo.deleteErr = errors.New("row lock timeout")

	err := svc.DeletePost(context.Background(), 1, 1)

	assert.Error(t, err)
	assert.Contains(t, postRepo.posts, 1, "пост не должен удаляться, если не удалось удалить его комментарии")
}

func TestPostService_DeletePost_RepoDeleteError_ReturnsError(t *testing.T) {
	svc, postRepo, _, _ := newTestPostService()
	postRepo.posts[1] = &model.Post{ID: 1, AuthorID: 1}
	postRepo.deleteErr = errors.New("disk full")

	err := svc.DeletePost(context.Background(), 1, 1)

	assert.Error(t, err)
}

// TestPostService_DeletePost_CommentDeletionExceedsIterationLimit проверяет защиту от бесконечного
// цикла в deleteAllCommentsForPost: если Delete возвращает nil, реально не удаляя комментарий
// (см. commentRepo.deleteNoOp), GetByPostID на следующей итерации снова видит ту же самую строку —
// без предела по количеству итераций цикл продолжался бы вечно
//
// commentDeletionPageSize и maxCommentDeletionIterations временно уменьшены именно для этого теста
// (через defer возвращаются обратно) - иначе пришлось бы реально прогонять цикл 100000 раз
// только ради того, чтобы проверить, что предел вообще есть
func TestPostService_DeletePost_CommentDeletionExceedsIterationLimit(t *testing.T) {
	origPageSize := commentDeletionPageSize
	origMaxIter := maxCommentDeletionIterations
	commentDeletionPageSize = 10
	maxCommentDeletionIterations = 3
	defer func() {
		commentDeletionPageSize = origPageSize
		maxCommentDeletionIterations = origMaxIter
	}()

	svc, postRepo, commentRepo, _ := newTestPostService()
	postRepo.posts[1] = &model.Post{ID: 1, AuthorID: 1}
	commentRepo.comments[1] = &model.Comment{ID: 1, PostID: 1}
	commentRepo.deleteNoOp = true // Delete сообщает об успехе, но ничего не удаляет

	err := svc.DeletePost(context.Background(), 1, 1)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete comments for post")
	// Пост не должен считаться удалённым, если очистка его комментариев
	// не завершилась успехом — postRepo.Delete вообще не должен вызываться
	assert.Contains(t, postRepo.posts, 1)
}

func TestPostService_PublishScheduledPosts_PublishesOnlyDueOnes(t *testing.T) {
	svc, postRepo, _, _ := newTestPostService()
	past := time.Now().Add(-1 * time.Hour)
	future := time.Now().Add(1 * time.Hour)
	postRepo.posts[1] = &model.Post{ID: 1, Status: model.PostStatusDraft, PublishAt: &past}
	postRepo.posts[2] = &model.Post{ID: 2, Status: model.PostStatusDraft, PublishAt: &future}

	published, err := svc.PublishScheduledPosts(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 1, published)
	assert.Equal(t, []int{1}, postRepo.publishedIDs)
}

func TestPostService_PublishScheduledPosts_GetScheduledPostsError_ReturnsError(t *testing.T) {
	svc, postRepo, _, _ := newTestPostService()
	postRepo.getScheduledErr = errors.New("db unavailable")

	_, err := svc.PublishScheduledPosts(context.Background())

	assert.Error(t, err)
}

// TestPostService_PublishScheduledPosts_PublishPostError_ContinuesWithRemaining — переписан:
// раньше ошибка публикации ОДНОГО поста прерывала весь цикл, и уже готовые к публикации посты
// после него оставались черновиками до следующего тика планировщика
// Теперь падение одного поста не должно мешать опубликовать остальные, готовые к публикации
//
// Сценарий: три поста, все три реально готовы к публикации (ShouldPublishNow == true у всех),
// но пост с ID=2 падает при PublishPost
// Ожидаем: посты 1 и 3 опубликованы (published == 2, оба ID в publishedIDs), пост 2 — нет, а
// возвращённая ошибка через errors.Is размечена именно про пост 2
func TestPostService_PublishScheduledPosts_PublishPostError_ContinuesWithRemaining(t *testing.T) {
	svc, postRepo, _, _ := newTestPostService()
	past := time.Now().Add(-1 * time.Hour)
	postRepo.posts[1] = &model.Post{ID: 1, Status: model.PostStatusDraft, PublishAt: &past}
	postRepo.posts[2] = &model.Post{ID: 2, Status: model.PostStatusDraft, PublishAt: &past}
	postRepo.posts[3] = &model.Post{ID: 3, Status: model.PostStatusDraft, PublishAt: &past}

	publishErr := errors.New("update failed")
	postRepo.publishErrByID = map[int]error{2: publishErr}

	published, err := svc.PublishScheduledPosts(context.Background())

	assert.Equal(t, 2, published, "посты 1 и 3 должны опубликоваться несмотря на сбой поста 2")
	require.Error(t, err)
	assert.ErrorIs(t, err, publishErr, "объединённая ошибка должна содержать исходную ошибку по посту 2")
	assert.ElementsMatch(t, []int{1, 3}, postRepo.publishedIDs)
}

// TestPostService_PublishScheduledPosts_AllPublishPostErrors_ReturnsZeroPublished —
// граничный случай предыдущего теста: если падают ВСЕ посты, published должен
// быть 0, а не частично ложным нулем, полученным по случайности
func TestPostService_PublishScheduledPosts_AllPublishPostErrors_ReturnsZeroPublished(t *testing.T) {
	svc, postRepo, _, _ := newTestPostService()
	past := time.Now().Add(-1 * time.Hour)
	postRepo.posts[1] = &model.Post{ID: 1, Status: model.PostStatusDraft, PublishAt: &past}
	postRepo.publishErr = errors.New("db unavailable")

	published, err := svc.PublishScheduledPosts(context.Background())

	assert.Equal(t, 0, published)
	require.Error(t, err)
}
