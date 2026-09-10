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

func TestPostService_PublishScheduledPosts_PublishPostError_ReturnsError(t *testing.T) {
	svc, postRepo, _, _ := newTestPostService()
	past := time.Now().Add(-1 * time.Hour)
	postRepo.posts[1] = &model.Post{ID: 1, Status: model.PostStatusDraft, PublishAt: &past}
	postRepo.publishErr = errors.New("update failed")

	_, err := svc.PublishScheduledPosts(context.Background())

	assert.Error(t, err)
}
