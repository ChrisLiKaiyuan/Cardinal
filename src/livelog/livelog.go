package livelog

import (
	"context"
	"errors"
	"sync"
)

type streamer struct {
	sync.Mutex

	streams map[int64]*stream
}

var errStreamNotFound = errors.New("stream: not found")

// New returns a new in-memory log streamer.
func New() *streamer {
	return &streamer{
		streams: make(map[int64]*stream),
	}
}

// Create adds a new log stream.
func (s *streamer) Create(id int64) error {
	s.Lock()
	s.streams[id] = newStream()
	s.Unlock()
	return nil
}

// Delete removes a log by id.
func (s *streamer) Delete(id int64) error {
	s.Lock()
	stream, ok := s.streams[id]
	if ok {
		delete(s.streams, id)
	}
	s.Unlock()
	if !ok {
		return errStreamNotFound
	}
	return stream.close()
}

// Write adds a new line into stream.
func (s *streamer) Write(id int64, line *Line) error {
	s.Lock()
	stream, ok := s.streams[id]
	s.Unlock()
	if !ok {
		return errStreamNotFound
	}
	return stream.write(line)
}

// Tail returns the end signal.
func (s *streamer) Tail(ctx context.Context, id int64) (<-chan *Line, <-chan error) {
	s.Lock()
	stream, ok := s.streams[id]
	s.Unlock()
	if !ok {
		return nil, nil
	}
	return stream.subscribe(ctx)
}

// Info returns the count of subscribers in each stream.
func (s *streamer) Info() map[int64]int {
	s.Lock()
	defer s.Unlock()
	info := map[int64]int{}
	for id, stream := range s.streams {
		stream.Lock()
		info[id] = len(stream.sub)
		stream.Unlock()
	}
	return info
}