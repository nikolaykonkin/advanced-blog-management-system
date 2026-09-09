package logger

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActionLogger_WritesEventAndFlushesOnClose(t *testing.T) {
	path := t.TempDir() + "/log.txt"

	l, err := New(path)
	require.NoError(t, err)

	start := time.Now()
	l.Log("user 5 created post 12")
	l.Log("user 5 created comment 7")

	// Close должен дождаться, пока воркер вычитает и запишет оба события
	// (включая задержку 1-2с на каждое), прежде чем вернуть управление
	require.NoError(t, l.Close())
	elapsed := time.Since(start)

	// Два события по 1-2с каждое -> минимум ~2с суммарно
	assert.GreaterOrEqual(t, elapsed, 2*time.Second)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	assert.Contains(t, string(data), "user 5 created post 12")
	assert.Contains(t, string(data), "user 5 created comment 7")
}

func TestActionLogger_NoEventsClosesImmediately(t *testing.T) {
	path := t.TempDir() + "/log.txt"

	l, err := New(path)
	require.NoError(t, err)

	start := time.Now()
	require.NoError(t, l.Close())

	assert.Less(t, time.Since(start), 500*time.Millisecond)
}

func TestActionLogger_InvalidPath_ReturnsError(t *testing.T) {
	// Директория "does-not-exist" не создаётся автоматически -
	// os.OpenFile должен вернуть ошибку, а New - обернуть ее
	_, err := New("/does-not-exist/log.txt")

	assert.Error(t, err)
}
