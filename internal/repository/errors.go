package repository

import "errors"

// Sentinel-ошибки репозиториев - возвращаются из Update/Delete/PublishPost,
// когда ни одна строка не затронута (rows == 0) — как правило, при race condition:
// запись удалили между GetByID и UPDATE/DELETE
// Сервисы маппят их в apperrors.ErrXxxNotFound, чтобы HTTP-слой отдавал 404, а не 500
var (
	ErrUserNotFound    = errors.New("user not found")
	ErrPostNotFound    = errors.New("post not found")
	ErrCommentNotFound = errors.New("comment not found")
)
