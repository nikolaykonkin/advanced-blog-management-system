// package logger реализует отложенное логирование действий пользователя
// в файл через канал и фоновую горутину-воркер
package logger

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"
)

const (
	eventBufferSize = 256

	minWriteDelay = 1 * time.Second
	maxWriteDelay = 2 * time.Second
)

// ActionLogger асинхронно записывает события действий пользователя в файл: события отправляются
// в канал, а фоновая горутина-воркер записывает их по одному с задержкой 1-2 секунды
type ActionLogger struct {
	events chan string
	file   *os.File
	wg     sync.WaitGroup
}

// New открывает (или создаёт) файл path в режиме дозаписи и запускаетворкер
// Close нужно вызвать при остановке сервера, чтобы воркер корректно завершился
func New(path string) (*ActionLogger, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("logger: failed to open log file %q: %w", path, err)
	}

	l := &ActionLogger{
		events: make(chan string, eventBufferSize),
		file:   file,
	}

	l.wg.Add(1)
	go l.worker()

	return l, nil
}

// Log отправляет событие в канал для отложенной записи
func (l *ActionLogger) Log(event string) {
	l.events <- event
}

// worker читает события из канала, выдерживает задержку 1-2 секунды и дописывает событие в файл
// Завершается, когда канал закрыт и все накопленные события вычитаны
func (l *ActionLogger) worker() {
	defer l.wg.Done()

	for event := range l.events {
		time.Sleep(randomDelay())

		line := fmt.Sprintf("[%s] %s\n", time.Now().UTC().Format(time.RFC3339), event)
		if _, err := l.file.WriteString(line); err != nil {
			fmt.Fprintf(os.Stderr, "logger: failed to write event %q: %v\n", event, err)
		}
	}
}

// Close закрывает канал событий, дожидается завершения воркера и закрывает файл
func (l *ActionLogger) Close() error {
	close(l.events)
	l.wg.Wait()
	return l.file.Close()
}

// randomDelay возвращает случайную задержку в диапазоне [1с, 2с]
func randomDelay() time.Duration {
	span := maxWriteDelay - minWriteDelay
	return minWriteDelay + time.Duration(rand.Int63n(int64(span)))
}
