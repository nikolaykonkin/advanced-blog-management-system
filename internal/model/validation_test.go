package model

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserCreateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     UserCreateRequest
		wantErr bool
	}{
		{"valid request", UserCreateRequest{Username: "user1", Email: "user@test.com", Password: "password123"}, false},
		{"empty username", UserCreateRequest{Username: "", Email: "user@test.com", Password: "password123"}, true},
		{"username too short", UserCreateRequest{Username: "ab", Email: "user@test.com", Password: "password123"}, true},
		{"username too long", UserCreateRequest{Username: strings.Repeat("a", 51), Email: "user@test.com", Password: "password123"}, true},
		{"invalid email", UserCreateRequest{Username: "user1", Email: "not-an-email", Password: "password123"}, true},
		{"empty password", UserCreateRequest{Username: "user1", Email: "user@test.com", Password: ""}, true},
		{"password too short", UserCreateRequest{Username: "user1", Email: "user@test.com", Password: "12345"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserLoginRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     UserLoginRequest
		wantErr bool
	}{
		{"valid request", UserLoginRequest{Email: "user@test.com", Password: "password123"}, false},
		{"empty email", UserLoginRequest{Email: "", Password: "password123"}, true},
		{"invalid email", UserLoginRequest{Email: "not-an-email", Password: "password123"}, true},
		{"empty password", UserLoginRequest{Email: "user@test.com", Password: ""}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPostCreateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     PostCreateRequest
		wantErr bool
	}{
		{"valid request", PostCreateRequest{Title: "Title", Content: "Content"}, false},
		{"empty title", PostCreateRequest{Title: "", Content: "Content"}, true},
		{"title too long", PostCreateRequest{Title: strings.Repeat("a", 201), Content: "Content"}, true},
		{"empty content", PostCreateRequest{Title: "Title", Content: ""}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPostUpdateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     PostUpdateRequest
		wantErr bool
	}{
		{"valid request", PostUpdateRequest{Title: "Title", Content: "Content"}, false},
		{"empty title", PostUpdateRequest{Title: "", Content: "Content"}, true},
		{"title too long", PostUpdateRequest{Title: strings.Repeat("a", 201), Content: "Content"}, true},
		{"empty content", PostUpdateRequest{Title: "Title", Content: ""}, true},
		{"empty status is allowed (omitempty)", PostUpdateRequest{Title: "Title", Content: "Content", Status: ""}, false},
		{"status draft is valid", PostUpdateRequest{Title: "Title", Content: "Content", Status: PostStatusDraft}, false},
		{"status published is valid", PostUpdateRequest{Title: "Title", Content: "Content", Status: PostStatusPublished}, false},
		{"unknown status is rejected", PostUpdateRequest{Title: "Title", Content: "Content", Status: "garbage"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCommentCreateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     CommentCreateRequest
		wantErr bool
	}{
		{"valid request", CommentCreateRequest{Content: "Nice post!"}, false},
		{"empty content", CommentCreateRequest{Content: ""}, true},
		{"content too long", CommentCreateRequest{Content: strings.Repeat("a", 1001)}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCommentUpdateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     CommentUpdateRequest
		wantErr bool
	}{
		{"valid request", CommentUpdateRequest{Content: "Updated comment"}, false},
		{"empty content", CommentUpdateRequest{Content: ""}, true},
		{"content too long", CommentUpdateRequest{Content: strings.Repeat("a", 1001)}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
