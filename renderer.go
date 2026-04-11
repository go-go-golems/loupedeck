package loupedeck

import (
	"errors"
	"sort"
	"sync"
	"time"
)

// RenderOptions control the full-B invalidation/coalescing scheduler.
type RenderOptions struct {
	FlushInterval time.Duration
	QueueSize     int
}

// DefaultRenderOptions provide modest coalescing without introducing long UI lag.
var DefaultRenderOptions = RenderOptions{
	FlushInterval: 40 * time.Millisecond,
	QueueSize:     256,
}

// RenderStats expose coarse invalidation/coalescing metrics.
type RenderStats struct {
	Invalidations         int
	CoalescedReplacements int
	FlushedCommands       int
	MaxPendingRegionCount int
}

type renderRequest struct {
	key string
	cmd outboundCommand
}

type renderScheduler struct {
	writer     *outboundWriter
	interval   time.Duration
	invalidate chan renderRequest
	closeCh    chan struct{}
	closeOnce  sync.Once

	mu    sync.Mutex
	stats RenderStats
}

func newRenderScheduler(writer *outboundWriter, opts RenderOptions) *renderScheduler {
	if opts.FlushInterval <= 0 {
		opts.FlushInterval = DefaultRenderOptions.FlushInterval
	}
	if opts.QueueSize <= 0 {
		opts.QueueSize = DefaultRenderOptions.QueueSize
	}
	r := &renderScheduler{
		writer:     writer,
		interval:   opts.FlushInterval,
		invalidate: make(chan renderRequest, opts.QueueSize),
		closeCh:    make(chan struct{}),
	}
	go r.loop()
	return r
}

func (r *renderScheduler) Close() {
	r.closeOnce.Do(func() {
		close(r.closeCh)
	})
}

func (r *renderScheduler) Stats() RenderStats {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.stats
}

func (r *renderScheduler) Invalidate(key string, cmd outboundCommand) error {
	select {
	case <-r.closeCh:
		return errors.New("render scheduler closed")
	case r.invalidate <- renderRequest{key: key, cmd: cmd}:
		return nil
	}
}

func (r *renderScheduler) loop() {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	pending := map[string]outboundCommand{}
	for {
		select {
		case <-r.closeCh:
			return
		case req := <-r.invalidate:
			r.recordInvalidation(false)
			if _, ok := pending[req.key]; ok {
				r.recordCoalesced()
			}
			pending[req.key] = req.cmd
			r.recordPendingCount(len(pending))
		case <-ticker.C:
			if len(pending) == 0 {
				continue
			}
			keys := make([]string, 0, len(pending))
			for k := range pending {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			toFlush := make([]outboundCommand, 0, len(keys))
			for _, k := range keys {
				toFlush = append(toFlush, pending[k])
				delete(pending, k)
			}
			for _, cmd := range toFlush {
				if err := r.writer.enqueue(cmd); err != nil {
					continue
				}
				r.recordFlushed()
			}
		}
	}
}

func (r *renderScheduler) recordInvalidation(_ bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stats.Invalidations++
}

func (r *renderScheduler) recordCoalesced() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stats.CoalescedReplacements++
}

func (r *renderScheduler) recordFlushed() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stats.FlushedCommands++
}

func (r *renderScheduler) recordPendingCount(count int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if count > r.stats.MaxPendingRegionCount {
		r.stats.MaxPendingRegionCount = count
	}
}
