package model

import (
	"time"
)

const (
	PostStatusDraft     = "draft"
	PostStatusPublished = "published"
)

// User представляет модель пользователя в системе
type User struct {
	ID        int       `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Post представляет модель поста в блоге
type Post struct {
	ID        int        `json:"id" db:"id"`
	Title     string     `json:"title" db:"title"`
	Content   string     `json:"content" db:"content"`
	AuthorID  int        `json:"author_id" db:"author_id"`
	Status    string     `json:"status" db:"status"`
	PublishAt *time.Time `json:"publish_at,omitempty" db:"publish_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// Comment представляет модель комментария к посту
type Comment struct {
	ID        int       `json:"id" db:"id"`
	Content   string    `json:"content" db:"content"`
	PostID    int       `json:"post_id" db:"post_id"`
	AuthorID  int       `json:"author_id" db:"author_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserCreateRequest представляет запрос на создание пользователя
type UserCreateRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// UserLoginRequest представляет запрос на вход пользователя
type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// PostCreateRequest представляет запрос на создание поста
type PostCreateRequest struct {
	Title     string     `json:"title" validate:"required,min=1,max=200"`
	Content   string     `json:"content" validate:"required,min=1"`
	PublishAt *time.Time `json:"publish_at,omitempty"`
}

// PostUpdateRequest представляет запрос на обновление поста
type PostUpdateRequest struct {
	Title     string     `json:"title" validate:"required,min=1,max=200"`
	Content   string     `json:"content" validate:"required,min=1"`
	Status    string     `json:"status,omitempty"`
	PublishAt *time.Time `json:"publish_at,omitempty"`
}

// CommentCreateRequest представляет запрос на создание комментария
type CommentCreateRequest struct {
	Content string `json:"content" validate:"required,min=1,max=1000"`
	PostID  int    `json:"post_id" validate:"required,gt=0"`
}

// CommentUpdateRequest представляет запрос на обновление комментария
type CommentUpdateRequest struct {
	Content string `json:"content" validate:"required,min=1,max=1000"`
}

// UserResponse - структура для ответа с данными пользователя (без пароля)
type UserResponse struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// TokenResponse - структура для ответа с JWT токеном
type TokenResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      UserResponse `json:"user"`
}

// PostResponse - структура для ответа с данными поста
type PostResponse struct {
	ID        int          `json:"id"`
	Title     string       `json:"title"`
	Content   string       `json:"content"`
	Author    UserResponse `json:"author"`
	Status    string       `json:"status"`
	PublishAt *time.Time   `json:"publish_at,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// CommentResponse - структура для ответа с данными комментария
type CommentResponse struct {
	ID        int          `json:"id"`
	Content   string       `json:"content"`
	PostID    int          `json:"post_id"`
	Author    UserResponse `json:"author"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// TODO: Реализовать метод ToResponse для User
// Преобразует User в UserResponse (без пароля)
// Должен скопировать поля ID, Username, Email, CreatedAt в новую структуру UserResponse
func (u *User) ToResponse() UserResponse {
	// TODO: реализовать
	return UserResponse{}
}

// TODO: Реализовать метод CanBeEditedBy для Post
// Проверяет может ли пользователь с userID редактировать этот пост
// Пост может редактировать только его автор
func (p *Post) CanBeEditedBy(userID int) bool {
	// TODO: реализовать - сравнить p.AuthorID с userID
	return false
}

// TODO: Реализовать метод CanBeDeletedBy для Post
// Проверяет может ли пользователь с userID удалить этот пост
// Пост может удалить только его автор
func (p *Post) CanBeDeletedBy(userID int) bool {
	// TODO: реализовать - сравнить p.AuthorID с userID
	return false
}

// TODO: Реализовать метод CanBeEditedBy для Comment
// Проверяет может ли пользователь с userID редактировать этот комментарий
// Комментарий может редактировать только его автор
func (c *Comment) CanBeEditedBy(userID int) bool {
	// TODO: реализовать - сравнить c.AuthorID с userID
	return false
}

// TODO: Реализовать метод CanBeDeletedBy для Comment
// Проверяет может ли пользователь с userID удалить этот комментарий
// Комментарий может удалить только его автор
func (c *Comment) CanBeDeletedBy(userID int) bool {
	// TODO: реализовать - сравнить c.AuthorID с userID
	return false
}

// TODO: Реализовать метод IsScheduled для Post
// Проверяет является ли пост отложенной публикацией (scheduled post)
// Пост считается scheduled если:
// - Status == PostStatusDraft (это черновик)
// - PublishAt != nil (время публикации установлено)
// - PublishAt находится в будущем (PublishAt.After(time.Now()))
func (p *Post) IsScheduled() bool {
	// TODO: реализовать проверку трех условий выше
	return false
}

// TODO: Реализовать метод ShouldPublishNow для Post
// Проверяет должен ли пост быть опубликован прямо сейчас
// Пост должен быть опубликован если:
// - Status == PostStatusDraft (это черновик)
// - PublishAt != nil (время публикации установлено)
// - PublishAt <= now (время публикации пришло или прошло)
func (p *Post) ShouldPublishNow() bool {
	// TODO: реализовать - проверка условий выше
	// Подсказка: используйте !PublishAt.After(time.Now()) для проверки что время прошло
	return false
}

// TODO: Реализовать метод Validate для UserCreateRequest
// Валидирует структуру используя github.com/go-playground/validator
// Должен вернуть nil если валидация прошла, или ошибку если нет
func (r *UserCreateRequest) Validate() error {
	// TODO: реализовать - создать validator и вызвать Struct(r)
	return nil
}

// TODO: Реализовать метод Validate для UserLoginRequest
// Валидирует структуру используя validator
func (r *UserLoginRequest) Validate() error {
	// TODO: реализовать
	return nil
}

// TODO: Реализовать метод Validate для PostCreateRequest
// Валидирует структуру используя validator
func (r *PostCreateRequest) Validate() error {
	// TODO: реализовать
	return nil
}

// TODO: Реализовать метод Validate для PostUpdateRequest
// Валидирует структуру используя validator
func (r *PostUpdateRequest) Validate() error {
	// TODO: реализовать
	return nil
}

// TODO: Реализовать метод Validate для CommentCreateRequest
// Валидирует структуру используя validator
func (r *CommentCreateRequest) Validate() error {
	// TODO: реализовать
	return nil
}

// TODO: Реализовать метод Validate для CommentUpdateRequest
// Валидирует структуру используя validator
func (r *CommentUpdateRequest) Validate() error {
	// TODO: реализовать
	return nil
}
