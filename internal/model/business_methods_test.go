package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestUser_ToResponse проверяет, что ToResponse копирует публичные поля User в UserResponse
// UserResponse как тип вообще не содержит поля Password - это гарантия на уровне типов, а не поведения;
// тест документирует само это намерение, а не проверяет отсутствие поля рефлексией
func TestUser_ToResponse(t *testing.T) {
	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	user := User{
		ID:        1,
		Username:  "nikolay",
		Email:     "nikolay@example.com",
		Password:  "super-secret-hash",
		CreatedAt: createdAt,
	}

	resp := user.ToResponse()

	assert.Equal(t, user.ID, resp.ID)
	assert.Equal(t, user.Username, resp.Username)
	assert.Equal(t, user.Email, resp.Email)
	assert.Equal(t, user.CreatedAt, resp.CreatedAt)
}

func TestPost_CanBeEditedBy(t *testing.T) {
	post := Post{AuthorID: 1}

	assert.True(t, post.CanBeEditedBy(1))
	assert.False(t, post.CanBeEditedBy(2))
}

func TestPost_CanBeDeletedBy(t *testing.T) {
	post := Post{AuthorID: 1}

	assert.True(t, post.CanBeDeletedBy(1))
	assert.False(t, post.CanBeDeletedBy(2))
}

func TestComment_CanBeEditedBy(t *testing.T) {
	comment := Comment{AuthorID: 1}

	assert.True(t, comment.CanBeEditedBy(1))
	assert.False(t, comment.CanBeEditedBy(2))
}

func TestComment_CanBeDeletedBy(t *testing.T) {
	comment := Comment{AuthorID: 1}

	assert.True(t, comment.CanBeDeletedBy(1))
	assert.False(t, comment.CanBeDeletedBy(2))
}

func TestPost_IsScheduled(t *testing.T) {
	future := time.Now().Add(1 * time.Hour)
	past := time.Now().Add(-1 * time.Hour)

	tests := []struct {
		name string
		post Post
		want bool
	}{
		{"draft with future publish_at is scheduled", Post{Status: PostStatusDraft, PublishAt: &future}, true},
		{"draft with past publish_at is not scheduled", Post{Status: PostStatusDraft, PublishAt: &past}, false},
		{"draft without publish_at is not scheduled", Post{Status: PostStatusDraft, PublishAt: nil}, false},
		{"published post is never scheduled", Post{Status: PostStatusPublished, PublishAt: &future}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.post.IsScheduled())
		})
	}
}

func TestPost_ShouldPublishNow(t *testing.T) {
	future := time.Now().Add(1 * time.Hour)
	past := time.Now().Add(-1 * time.Hour)

	tests := []struct {
		name string
		post Post
		want bool
	}{
		{"draft with past publish_at should publish", Post{Status: PostStatusDraft, PublishAt: &past}, true},
		{"draft with future publish_at should not publish yet", Post{Status: PostStatusDraft, PublishAt: &future}, false},
		{"draft without publish_at should not publish", Post{Status: PostStatusDraft, PublishAt: nil}, false},
		{"already published post should not publish again", Post{Status: PostStatusPublished, PublishAt: &past}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.post.ShouldPublishNow())
		})
	}
}
