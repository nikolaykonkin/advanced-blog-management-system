package service

import (
	"context"
	"fmt"
	"testing"

	"advanced-blog-management-system/internal/errors/apperrors"
	"advanced-blog-management-system/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCommentService() (*CommentService, *fakePostRepository, *fakeCommentRepository, *fakeActionLogger) {
	postRepo := newFakePostRepository()
	commentRepo := newFakeCommentRepository()
	userRepo := newFakeUserRepository()
	logger := &fakeActionLogger{}
	svc := NewCommentService(commentRepo, postRepo, userRepo, logger)
	return svc, postRepo, commentRepo, logger
}

// postID приходит отдельным параметром из URL (/posts/{id}/comments),
// а не из тела запроса — CommentCreateRequest его не содержит

func TestCommentService_CreateComment_Success(t *testing.T) {
	svc, postRepo, _, _ := newTestCommentService()
	postRepo.posts[1] = &model.Post{ID: 1, Status: model.PostStatusPublished}

	comment, err := svc.CreateComment(context.Background(), &model.CommentCreateRequest{Content: "Nice post!"}, 1, 5)

	require.NoError(t, err)
	assert.Equal(t, "Nice post!", comment.Content)
	assert.Equal(t, 1, comment.PostID)
	assert.Equal(t, 5, comment.AuthorID)
}

func TestCommentService_CreateComment_PostNotFound_ReturnsErrPostNotFound(t *testing.T) {
	svc, _, _, _ := newTestCommentService()

	_, err := svc.CreateComment(context.Background(), &model.CommentCreateRequest{Content: "Nice post!"}, 999, 5)

	assert.ErrorIs(t, err, apperrors.ErrPostNotFound)
}

func TestCommentService_CreateComment_PostNotPublished_ReturnsErrPostNotPublished(t *testing.T) {
	svc, postRepo, _, _ := newTestCommentService()
	postRepo.posts[1] = &model.Post{ID: 1, Status: model.PostStatusDraft}

	_, err := svc.CreateComment(context.Background(), &model.CommentCreateRequest{Content: "Nice post!"}, 1, 5)

	assert.ErrorIs(t, err, ErrPostNotPublished)
}

func TestCommentService_CreateComment_InvalidRequest_ReturnsValidationError(t *testing.T) {
	svc, postRepo, _, _ := newTestCommentService()
	postRepo.posts[1] = &model.Post{ID: 1, Status: model.PostStatusPublished}

	_, err := svc.CreateComment(context.Background(), &model.CommentCreateRequest{Content: ""}, 1, 5)

	assert.Error(t, err)
}

func TestCommentService_CreateComment_LogsActionEvent(t *testing.T) {
	svc, postRepo, _, logger := newTestCommentService()
	postRepo.posts[1] = &model.Post{ID: 1, Status: model.PostStatusPublished}

	comment, err := svc.CreateComment(context.Background(), &model.CommentCreateRequest{Content: "Nice post!"}, 1, 7)
	require.NoError(t, err)

	require.Len(t, logger.events, 1)
	assert.Equal(t, fmt.Sprintf("user 7 created comment %d", comment.ID), logger.events[0])
}

func TestCommentService_GetComment_NotFound_ReturnsErrCommentNotFound(t *testing.T) {
	svc, _, _, _ := newTestCommentService()

	_, err := svc.GetComment(context.Background(), 999)

	assert.ErrorIs(t, err, apperrors.ErrCommentNotFound)
}

func TestCommentService_GetCommentsByPostID_PostNotFound_ReturnsErrPostNotFound(t *testing.T) {
	svc, _, _, _ := newTestCommentService()

	_, err := svc.GetCommentsByPostID(context.Background(), 999, 10, 0)

	assert.ErrorIs(t, err, apperrors.ErrPostNotFound)
}

func TestCommentService_GetCommentsByPostID_Success(t *testing.T) {
	svc, postRepo, commentRepo, _ := newTestCommentService()
	postRepo.posts[1] = &model.Post{ID: 1, Status: model.PostStatusPublished}
	commentRepo.comments[1] = &model.Comment{ID: 1, PostID: 1, Content: "First"}
	commentRepo.comments[2] = &model.Comment{ID: 2, PostID: 2, Content: "Other post"}

	comments, err := svc.GetCommentsByPostID(context.Background(), 1, 10, 0)

	require.NoError(t, err)
	require.Len(t, comments, 1)
	assert.Equal(t, "First", comments[0].Content)
}

func TestCommentService_UpdateComment_NotOwner_ReturnsErrForbidden(t *testing.T) {
	svc, _, commentRepo, _ := newTestCommentService()
	commentRepo.comments[1] = &model.Comment{ID: 1, AuthorID: 42, Content: "Old"}

	_, err := svc.UpdateComment(context.Background(), 1, &model.CommentUpdateRequest{Content: "New"}, 1)

	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestCommentService_UpdateComment_NotFound_ReturnsErrCommentNotFound(t *testing.T) {
	svc, _, _, _ := newTestCommentService()

	_, err := svc.UpdateComment(context.Background(), 999, &model.CommentUpdateRequest{Content: "New"}, 1)

	assert.ErrorIs(t, err, apperrors.ErrCommentNotFound)
}

func TestCommentService_UpdateComment_Owner_UpdatesContent(t *testing.T) {
	svc, _, commentRepo, _ := newTestCommentService()
	commentRepo.comments[1] = &model.Comment{ID: 1, AuthorID: 1, Content: "Old"}

	comment, err := svc.UpdateComment(context.Background(), 1, &model.CommentUpdateRequest{Content: "New"}, 1)

	require.NoError(t, err)
	assert.Equal(t, "New", comment.Content)
}

func TestCommentService_UpdateComment_InvalidRequest_ReturnsValidationError(t *testing.T) {
	svc, _, commentRepo, _ := newTestCommentService()
	commentRepo.comments[1] = &model.Comment{ID: 1, AuthorID: 1, Content: "Old"}

	_, err := svc.UpdateComment(context.Background(), 1, &model.CommentUpdateRequest{Content: ""}, 1)

	assert.Error(t, err)
}

func TestCommentService_DeleteComment_NotOwner_ReturnsErrForbidden(t *testing.T) {
	svc, _, commentRepo, _ := newTestCommentService()
	commentRepo.comments[1] = &model.Comment{ID: 1, AuthorID: 42}

	err := svc.DeleteComment(context.Background(), 1, 1)

	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestCommentService_DeleteComment_NotFound_ReturnsErrCommentNotFound(t *testing.T) {
	svc, _, _, _ := newTestCommentService()

	err := svc.DeleteComment(context.Background(), 999, 1)

	assert.ErrorIs(t, err, apperrors.ErrCommentNotFound)
}

func TestCommentService_DeleteComment_Owner_DeletesComment(t *testing.T) {
	svc, _, commentRepo, _ := newTestCommentService()
	commentRepo.comments[1] = &model.Comment{ID: 1, AuthorID: 1}

	err := svc.DeleteComment(context.Background(), 1, 1)

	require.NoError(t, err)
	assert.NotContains(t, commentRepo.comments, 1)
}
