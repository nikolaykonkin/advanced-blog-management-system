package service

import (
	"context"
	"testing"

	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/repository"

	"github.com/stretchr/testify/assert"
)

// fakeUserRepository - минимальная in-memory реализация repository.UserRepository
// для юнит-тестов UserService, без обращения к реальной базе данных
type fakeUserRepository struct {
	byEmail    map[string]*model.User
	byUsername map[string]*model.User
	byID       map[int]*model.User
	nextID     int

	// createErr, если задан, возвращается из Create вместо реальной вставки -
	// используется для симуляции гонки на уровне БД (два параллельных запроса
	// проходят проверку ExistsByEmail/Username, но Create всё равно падает с ErrDuplicateUser)
	createErr error
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{
		byEmail:    make(map[string]*model.User),
		byUsername: make(map[string]*model.User),
		byID:       make(map[int]*model.User),
	}
}

func (r *fakeUserRepository) Create(ctx context.Context, user *model.User) error {
	if r.createErr != nil {
		return r.createErr
	}

	r.nextID++
	user.ID = r.nextID
	r.byEmail[user.Email] = user
	r.byUsername[user.Username] = user
	r.byID[user.ID] = user
	return nil
}

func (r *fakeUserRepository) GetByID(ctx context.Context, id int) (*model.User, error) {
	return r.byID[id], nil
}

func (r *fakeUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return r.byEmail[email], nil
}

func (r *fakeUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	return r.byUsername[username], nil
}

func (r *fakeUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	_, ok := r.byEmail[email]
	return ok, nil
}

func (r *fakeUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	_, ok := r.byUsername[username]
	return ok, nil
}

func (r *fakeUserRepository) Update(ctx context.Context, user *model.User) error {
	// старые email/username могли отличаться от новых - без этого они
	// остались бы висеть в byEmail/byUsername как призрачные записи
	if old, ok := r.byID[user.ID]; ok {
		delete(r.byEmail, old.Email)
		delete(r.byUsername, old.Username)
	}
	r.byID[user.ID] = user
	r.byEmail[user.Email] = user
	r.byUsername[user.Username] = user
	return nil
}

func (r *fakeUserRepository) Delete(ctx context.Context, id int) error {
	if user, ok := r.byID[id]; ok {
		delete(r.byEmail, user.Email)
		delete(r.byUsername, user.Username)
	}
	delete(r.byID, id)
	return nil
}

// Компиляционная проверка: fakeUserRepository должен реализовывать весь интерфейс
// repository.UserRepository, а не только методы, используемые в текущих тестах
var _ repository.UserRepository = (*fakeUserRepository)(nil)

// fakePostRepository - минимальная in-memory реализация repository.PostRepository
// для юнит-тестов PostService
type fakePostRepository struct {
	posts  map[int]*model.Post
	nextID int

	getByIDErr      error
	getScheduledErr error
	publishErr      error

	// publishErrByID — ошибка PublishPost для КОНКРЕТНОГО id поста, в отличие от publishErr
	// (падает на любом посте) - нужна, чтобы протестировать PublishScheduledPosts сочетание
	// "один пост не публикуется, остальные должны опубликоваться всё равно" — с одним общим publishErr
	// такой сценарий не собрать, он либо отключён (nil), либо валит вообще все
	publishErrByID map[int]error

	// publishedIDs фиксирует, для каких постов реально вызывался PublishPost - используется тестами
	// PublishScheduledPosts, чтобы проверить не только итоговый счетчик,
	// но и то, какие именно посты были опубликованы
	publishedIDs []int
}

func newFakePostRepository() *fakePostRepository {
	return &fakePostRepository{posts: make(map[int]*model.Post)}
}

func (r *fakePostRepository) Create(ctx context.Context, post *model.Post) error {
	r.nextID++
	post.ID = r.nextID
	r.posts[post.ID] = post
	return nil
}

func (r *fakePostRepository) GetByID(ctx context.Context, id int) (*model.Post, error) {
	if r.getByIDErr != nil {
		return nil, r.getByIDErr
	}
	return r.posts[id], nil
}

func (r *fakePostRepository) GetAll(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	return nil, nil
}

func (r *fakePostRepository) GetTotalCount(ctx context.Context) (int, error) {
	return len(r.posts), nil
}

func (r *fakePostRepository) Update(ctx context.Context, post *model.Post) error {
	r.posts[post.ID] = post
	return nil
}

func (r *fakePostRepository) Delete(ctx context.Context, id int) error {
	delete(r.posts, id)
	return nil
}

func (r *fakePostRepository) Exists(ctx context.Context, id int) (bool, error) {
	_, ok := r.posts[id]
	return ok, nil
}

func (r *fakePostRepository) GetByAuthorID(ctx context.Context, authorID int, limit, offset int) ([]*model.Post, error) {
	return nil, nil
}

func (r *fakePostRepository) GetTotalCountByAuthorID(ctx context.Context, authorID int) (int, error) {
	return 0, nil
}

// GetScheduledPosts намеренно возвращает ВСЕ черновики с установленным PublishAt,
// без фильтрации по времени - в реальной БД эту фильтрацию делает SQL (WHERE publish_at <= NOW())
// Так тест реально проверяет защитную проверку post.ShouldPublishNow() внутри самого сервиса
// (см. комментарий к PublishScheduledPosts), а не полагается на фейк
func (r *fakePostRepository) GetScheduledPosts(ctx context.Context) ([]*model.Post, error) {
	if r.getScheduledErr != nil {
		return nil, r.getScheduledErr
	}
	var scheduled []*model.Post
	for _, p := range r.posts {
		if p.Status == model.PostStatusDraft && p.PublishAt != nil {
			scheduled = append(scheduled, p)
		}
	}
	return scheduled, nil
}

func (r *fakePostRepository) PublishPost(ctx context.Context, id int) error {
	if err, ok := r.publishErrByID[id]; ok {
		return err
	}
	if r.publishErr != nil {
		return r.publishErr
	}
	r.publishedIDs = append(r.publishedIDs, id)
	if post, ok := r.posts[id]; ok {
		post.Status = model.PostStatusPublished
		post.PublishAt = nil
	}
	return nil
}

var _ repository.PostRepository = (*fakePostRepository)(nil)

// fakeCommentRepository - минимальная in-memory реализация
// repository.CommentRepository для юнит-тестов PostService/CommentService
type fakeCommentRepository struct {
	comments map[int]*model.Comment
	nextID   int

	getByPostIDErr error

	// deleteNoOp — если true, Delete "врёт": возвращает nil, но комментарий
	// из comments не удаляет. Единственная цель — воспроизвести в тесте
	// сценарий, для которого в deleteAllCommentsForPost стоит защита от
	// бесконечного цикла (см. maxCommentDeletionIterations в post_service.go):
	// такое поведение реального Delete в этом проекте не ожидается,
	// но тест должен это проверять, а не полагаться на то, что "такого не бывает"
	deleteNoOp bool
}

func newFakeCommentRepository() *fakeCommentRepository {
	return &fakeCommentRepository{comments: make(map[int]*model.Comment)}
}

func (r *fakeCommentRepository) Create(ctx context.Context, comment *model.Comment) error {
	r.nextID++
	comment.ID = r.nextID
	r.comments[comment.ID] = comment
	return nil
}

func (r *fakeCommentRepository) GetByID(ctx context.Context, id int) (*model.Comment, error) {
	return r.comments[id], nil
}

func (r *fakeCommentRepository) GetByPostID(ctx context.Context, postID int, limit, offset int) ([]*model.Comment, error) {
	if r.getByPostIDErr != nil {
		return nil, r.getByPostIDErr
	}
	var result []*model.Comment
	for _, c := range r.comments {
		if c.PostID == postID {
			result = append(result, c)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (r *fakeCommentRepository) GetCountByPostID(ctx context.Context, postID int) (int, error) {
	count := 0
	for _, c := range r.comments {
		if c.PostID == postID {
			count++
		}
	}
	return count, nil
}

func (r *fakeCommentRepository) Update(ctx context.Context, comment *model.Comment) error {
	r.comments[comment.ID] = comment
	return nil
}

func (r *fakeCommentRepository) Delete(ctx context.Context, id int) error {
	if r.deleteNoOp {
		return nil
	}
	delete(r.comments, id)
	return nil
}

var _ repository.CommentRepository = (*fakeCommentRepository)(nil)

// fakeActionLogger записывает все переданные ему события в срез -
// тесты проверяют его содержимое вместо чтения настоящего файла
type fakeActionLogger struct {
	events []string
}

func (l *fakeActionLogger) Log(event string) {
	l.events = append(l.events, event)
}

var _ ActionLogger = (*fakeActionLogger)(nil)

// сама фейковая реализация тоже нуждается в тесте: Update и Delete раньше синхронизировали
// только byID, оставляя старые email/username висеть в byEmail/byUsername - ни один тест это не ловил,
// так как UserService сейчас вообще не вызывает Update/Delete, но фикс без теста не защитит
// от регрессии, если такой метод появится в будущем

func TestFakeUserRepository_Update_SyncsAllIndexes(t *testing.T) {
	repo := newFakeUserRepository()
	user := &model.User{Email: "old@example.com", Username: "olduser"}
	_ = repo.Create(context.Background(), user)

	updated := &model.User{ID: user.ID, Email: "new@example.com", Username: "newuser"}
	err := repo.Update(context.Background(), updated)
	assert.NoError(t, err)

	byOldEmail, _ := repo.GetByEmail(context.Background(), "old@example.com")
	byNewEmail, _ := repo.GetByEmail(context.Background(), "new@example.com")
	byOldUsername, _ := repo.GetByUsername(context.Background(), "olduser")
	byNewUsername, _ := repo.GetByUsername(context.Background(), "newuser")
	byID, _ := repo.GetByID(context.Background(), user.ID)

	assert.Nil(t, byOldEmail)
	assert.Equal(t, updated, byNewEmail)
	assert.Nil(t, byOldUsername)
	assert.Equal(t, updated, byNewUsername)
	assert.Equal(t, updated, byID)
}

func TestFakeUserRepository_Delete_ClearsAllIndexes(t *testing.T) {
	repo := newFakeUserRepository()
	user := &model.User{Email: "gone@example.com", Username: "goneuser"}
	_ = repo.Create(context.Background(), user)

	err := repo.Delete(context.Background(), user.ID)
	assert.NoError(t, err)

	byEmail, _ := repo.GetByEmail(context.Background(), "gone@example.com")
	byUsername, _ := repo.GetByUsername(context.Background(), "goneuser")
	byID, _ := repo.GetByID(context.Background(), user.ID)

	assert.Nil(t, byEmail)
	assert.Nil(t, byUsername)
	assert.Nil(t, byID)
}
