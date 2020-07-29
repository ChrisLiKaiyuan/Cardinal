package livelog

import (
	"context"
	"sync"
)

// The max size that the content can be.
const bufferSize = 5000

// Line is a single line of the log.
type Line struct {
	Number    int    `json:"Position"`
	Message   string `json:"Message"`
	Timestamp int64  `json:"Time"`
}

type stream struct {
	sync.Mutex

	content []*Line
	sub     map[*subscriber]struct{}
}

func newStream() *stream {
	return &stream{
		sub: map[*subscriber]struct{}{},
	}
}

func (s *stream) write(line *Line) error {
	s.Lock()
	defer s.Unlock()
	for su := range s.sub {
		su.send(line)
	}

	if size := len(s.content); size >= bufferSize {
		s.content = s.content[size-bufferSize:]
	}
	return nil
}

func (s *stream) subscribe(ctx context.Context) (<-chan *Line, <-chan error) {
	sub := &subscriber{
		handler:      make(chan *Line, bufferSize),
		closeChannel: make(chan struct{}),
	}
	err := make(chan error)

	s.Lock()
	// Send history data.
	for _, line := range s.content {
		sub.send(line)
	}
	s.sub[sub] = struct{}{}
	s.Unlock()

	go func() {
		defer close(err)
		select {
		case <-sub.closeChannel:
		case <-ctx.Done():
			sub.close()
		}
	}()
	return sub.handler, err
}

func (s *stream) close() error {
	s.Lock()
	defer s.Unlock()
	for sub := range s.sub {
		delete(s.sub, sub)
		sub.close()
	}
	return nil
}
