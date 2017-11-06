package store

import (
	"log"

	"context"
)

// The watchMgr keeps track of all the watches that are set on the store.
type watchMgr struct {
	// The mailbox is the interaction point of the manager.
	// The manager understands messages that result from these events:
	// - New watch issued on a specific bucket
	// - New write was performed on a specific bucket
	// - Specific watch is canceled by the client
	mailbox chan interface{}
	// watchersPerKey keeps track of all the watchers for specific keys.
	watchersPerKey map[string]map[uint64]chan WatchResponse
	// watchCounter is used to assign a unique ID to each watcher.
	// TODO: need to figure out a better way to do this, as this will eventually
	// overflow. Maybe use UUIDs?
	watchCounter uint64
	logger       *log.Logger
}

func (mgr *watchMgr) run(ctx context.Context) {
	for {
		select {
		case msg := <-mgr.mailbox:
			switch m := msg.(type) {
			case newWatchMsg:
				// stash the response channel in the right map so that we can use it
				// later for notifying on writes
				keyWatchers, ok := mgr.watchersPerKey[m.bucket]
				if !ok {
					keyWatchers = make(map[uint64]chan WatchResponse)
					mgr.watchersPerKey[m.bucket] = keyWatchers
				}
				keyWatchers[mgr.watchCounter] = m.respChan

				// Detect watch cancelation
				go func(id uint64) {
					<-m.ctx.Done()
					mgr.mailbox <- watchCanceledMsg{bucket: m.bucket, watcherID: id}
				}(mgr.watchCounter)

				mgr.watchCounter++

			case watchCanceledMsg:
				keyWatchers, ok := mgr.watchersPerKey[m.bucket]
				if !ok {
					continue
				}
				w, ok := keyWatchers[m.watcherID]
				if !ok {
					continue
				}
				close(w)
				delete(keyWatchers, m.watcherID)

			case writeOnKeyMsg:
				watchers, ok := mgr.watchersPerKey[m.bucket]
				if !ok {
					mgr.logger.Printf("cannot send key write notification, bucket %q is nil\n", m.bucket)
					continue
				}
				for _, w := range watchers {
					// Notify each client in a non-blocking fashion.
					// Dropping msgs is OK.
					select {
					case w <- WatchResponse{Key: m.key, Value: m.value}:
					default:
						mgr.logger.Printf("dropped notification\n")
					}
				}
			default:
				mgr.logger.Printf("unknown message sent to watch mgr: %T\n", msg)
			}
		case <-ctx.Done():
			close(mgr.mailbox)
			return
		}
	}
}

type newWatchMsg struct {
	bucket   string
	respChan chan WatchResponse
	ctx      context.Context
}

type writeOnKeyMsg struct {
	bucket string
	key    string
	value  []byte
}

type watchCanceledMsg struct {
	bucket    string
	watcherID uint64
}
