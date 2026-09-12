package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"advanced-blog-management-system/internal/middleware"
	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// минимальные in-memory фейки для сборки хендлеров - AuthHandler/CommentHandler принимают
// конкретные *service.XxxService, а не интерфейсы, поэтому фейки из internal/service недоступны отсюда
// (они неэкспортированы и лежат в _test.go другого пакета) и приходится заводить свои, отдельные

type fakeUserRepo struct {
	byID    map[int]*model.User
	byEmail map[string]*model.User
	nextID  int
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byID: map[int]*model.User{}, byEmail: map[string]*model.User{}}
}

func (r *fakeUserRepo) Create(ctx context.Context, u *model.User) error {
	r.nextID++
	u.ID = r.nextID
	r.byID[u.ID] = u
	r.byEmail[u.Email] = u
	return nil
}
func (r *fakeUserRepo) GetByID(ctx context.Context, id int) (*model.User, error) {
	return r.byID[id], nil
}
func (r *fakeUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return r.byEmail[email], nil
}
func (r *fakeUserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	return nil, nil
}
func (r *fakeUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	_, ok := r.byEmail[email]
	return ok, nil
}
func (r *fakeUserRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return false, nil
}
func (r *fakeUserRepo) Update(ctx context.Context, u *model.User) error { r.byID[u.ID] = u; return nil }
func (r *fakeUserRepo) Delete(ctx context.Context, id int) error        { delete(r.byID, id); return nil }

type fakePostRepo struct {
	posts  map[int]*model.Post
	nextID int
}

func newFakePostRepo() *fakePostRepo { return &fakePostRepo{posts: map[int]*model.Post{}} }

func (r *fakePostRepo) Create(ctx context.Context, p *model.Post) error {
	r.nextID++
	p.ID = r.nextID
	r.posts[p.ID] = p
	return nil
}
func (r *fakePostRepo) GetByID(ctx context.Context, id int) (*model.Post, error) {
	return r.posts[id], nil
}
func (r *fakePostRepo) GetAll(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	return nil, nil
}
func (r *fakePostRepo) GetTotalCount(ctx context.Context) (int, error) { return len(r.posts), nil }
func (r *fakePostRepo) Update(ctx context.Context, p *model.Post) error {
	r.posts[p.ID] = p
	return nil
}
func (r *fakePostRepo) Delete(ctx context.Context, id int) error { delete(r.posts, id); return nil }
func (r *fakePostRepo) Exists(ctx context.Context, id int) (bool, error) {
	_, ok := r.posts[id]
	return ok, nil
}
func (r *fakePostRepo) GetByAuthorID(ctx context.Context, authorID, limit, offset int) ([]*model.Post, error) {
	return nil, nil
}
func (r *fakePostRepo) GetTotalCountByAuthorID(ctx context.Context, authorID int) (int, error) {
	return 0, nil
}
func (r *fakePostRepo) GetScheduledPosts(ctx context.Context) ([]*model.Post, error) { return nil, nil }
func (r *fakePostRepo) PublishPost(ctx context.Context, id int) error                { return nil }

type fakeCommentRepo struct {
	comments map[int]*model.Comment
	nextID   int
}

func newFakeCommentRepo() *fakeCommentRepo {
	return &fakeCommentRepo{comments: map[int]*model.Comment{}}
}

func (r *fakeCommentRepo) Create(ctx context.Context, c *model.Comment) error {
	r.nextID++
	c.ID = r.nextID
	r.comments[c.ID] = c
	return nil
}
func (r *fakeCommentRepo) GetByID(ctx context.Context, id int) (*model.Comment, error) {
	return r.comments[id], nil
}
func (r *fakeCommentRepo) GetByPostID(ctx context.Context, postID, limit, offset int) ([]*model.Comment, error) {
	return nil, nil
}
func (r *fakeCommentRepo) GetCountByPostID(ctx context.Context, postID int) (int, error) {
	return 0, nil
}
func (r *fakeCommentRepo) Update(ctx context.Context, c *model.Comment) error {
	r.comments[c.ID] = c
	return nil
}
func (r *fakeCommentRepo) Delete(ctx context.Context, id int) error {
	delete(r.comments, id)
	return nil
}

type noopActionLogger struct{}

func (noopActionLogger) Log(event string) {}

// ---------------------------------------------------------------------
// HealthCheckHandler
// ---------------------------------------------------------------------

func TestHealthCheckHandler_ReturnsOK(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()

	HealthCheckHandler(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"status":"ok"}`, rec.Body.String())
}

// ---------------------------------------------------------------------
// AuthHandler
// ---------------------------------------------------------------------

func TestAuthHandler_Register_Success(t *testing.T) {
	h := NewAuthHandler(service.NewUserService(newFakeUserRepo()), "test-secret")

	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(
		`{"username":"johndoe","email":"john@example.com","password":"password123"}`))
	rec := httptest.NewRecorder()

	h.RegisterHandler(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	var resp model.TokenResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "john@example.com", resp.User.Email)
}

func TestAuthHandler_Register_InvalidBody_ReturnsBadRequest(t *testing.T) {
	h := NewAuthHandler(service.NewUserService(newFakeUserRepo()), "test-secret")

	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(`{"email":"not-an-email"}`))
	rec := httptest.NewRecorder()

	h.RegisterHandler(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthHandler_Register_DuplicateEmail_ReturnsConflict(t *testing.T) {
	h := NewAuthHandler(service.NewUserService(newFakeUserRepo()), "test-secret")
	body := `{"username":"johndoe","email":"john@example.com","password":"password123"}`
	h.RegisterHandler(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(body)))

	rec := httptest.NewRecorder()
	h.RegisterHandler(rec, httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(body)))

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	h := NewAuthHandler(service.NewUserService(newFakeUserRepo()), "test-secret")
	registerBody := `{"username":"johndoe","email":"john@example.com","password":"password123"}`
	h.RegisterHandler(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(registerBody)))

	rec := httptest.NewRecorder()
	h.LoginHandler(rec, httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(
		`{"email":"john@example.com","password":"password123"}`)))

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthHandler_Login_WrongPassword_ReturnsUnauthorized(t *testing.T) {
	h := NewAuthHandler(service.NewUserService(newFakeUserRepo()), "test-secret")
	registerBody := `{"username":"johndoe","email":"john@example.com","password":"password123"}`
	h.RegisterHandler(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(registerBody)))

	rec := httptest.NewRecorder()
	h.LoginHandler(rec, httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(
		`{"email":"john@example.com","password":"wrongpassword"}`)))

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ---------------------------------------------------------------------
// CommentHandler.CreateComment
// ---------------------------------------------------------------------

func newTestCommentHandler(postRepo *fakePostRepo) *CommentHandler {
	svc := service.NewCommentService(newFakeCommentRepo(), postRepo, newFakeUserRepo(), noopActionLogger{})
	return NewCommentHandler(svc)
}

// chi.URLParam читает параметр из RouteContext, который заполняется только реальным роутером -
// поэтому здесь используется chi.NewRouter(), а не вызов хендлера напрямую, как в тестах AuthHandler выше
func TestCommentHandler_CreateComment_Success(t *testing.T) {
	postRepo := newFakePostRepo()
	postRepo.posts[1] = &model.Post{ID: 1, Status: model.PostStatusPublished}
	router := chi.NewRouter()
	router.Post("/api/posts/{id}/comments", newTestCommentHandler(postRepo).CreateComment)

	req := httptest.NewRequest(http.MethodPost, "/api/posts/1/comments", strings.NewReader(`{"content":"Nice post!"}`))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserKey, 7))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestCommentHandler_CreateComment_Unauthenticated_ReturnsUnauthorized(t *testing.T) {
	postRepo := newFakePostRepo()
	postRepo.posts[1] = &model.Post{ID: 1, Status: model.PostStatusPublished}
	router := chi.NewRouter()
	router.Post("/api/posts/{id}/comments", newTestCommentHandler(postRepo).CreateComment)

	req := httptest.NewRequest(http.MethodPost, "/api/posts/1/comments", strings.NewReader(`{"content":"Nice post!"}`))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// пост существует, но еще черновик - CreateComment должен вернуть 400 через
// commentServiceErrorStatus(service.ErrPostNotPublished), а не 500
func TestCommentHandler_CreateComment_PostNotPublished_ReturnsBadRequest(t *testing.T) {
	postRepo := newFakePostRepo()
	postRepo.posts[1] = &model.Post{ID: 1, Status: model.PostStatusDraft}
	router := chi.NewRouter()
	router.Post("/api/posts/{id}/comments", newTestCommentHandler(postRepo).CreateComment)

	req := httptest.NewRequest(http.MethodPost, "/api/posts/1/comments", strings.NewReader(`{"content":"Nice post!"}`))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserKey, 7))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCommentHandler_CreateComment_InvalidPostID_ReturnsBadRequest(t *testing.T) {
	router := chi.NewRouter()
	router.Post("/api/posts/{id}/comments", newTestCommentHandler(newFakePostRepo()).CreateComment)

	req := httptest.NewRequest(http.MethodPost, "/api/posts/abc/comments", strings.NewReader(`{"content":"Nice post!"}`))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserKey, 7))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
