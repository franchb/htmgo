package ws

import (
	"fmt"
	"github.com/franchb/htmgo/extensions/websocket/internal/wsutil"
	"github.com/franchb/htmgo/extensions/websocket/session"
	"sync"
)

type MessageHandler struct {
	manager *wsutil.SocketManager
}

func NewMessageHandler(manager *wsutil.SocketManager) *MessageHandler {
	return &MessageHandler{manager: manager}
}

func (h *MessageHandler) OnServerSideEvent(e ServerSideEvent) {
	fmt.Printf("received server side event: %s\n", e.Event)
	hashes, ok := serverEventNamesToHash.Load(e.Event)
	if !ok {
		return
	}

	// Collect the hashes we need to invoke. If not broadcasting to everyone,
	// intersect with the hashes registered for this specific session.
	type handlerEntry struct {
		hash      KeyHash
		sessionId session.Id
		cb        Handler
	}

	var toRun []handlerEntry

	if e.SessionId != "*" {
		hashesForSession, ok2 := sessionIdToHashes.Load(e.SessionId)
		if ok2 {
			hashesForSession.Range(func(hash KeyHash, _ bool) bool {
				if _, found := hashes.Load(hash); found {
					if cb, cbOk := handlers.Load(hash); cbOk {
						if sid, sidOk := hashesToSessionId.Load(hash); sidOk {
							toRun = append(toRun, handlerEntry{hash: hash, sessionId: sid, cb: cb})
						}
					}
				}
				return true
			})
		}
	} else {
		hashes.Range(func(hash KeyHash, _ bool) bool {
			if cb, cbOk := handlers.Load(hash); cbOk {
				if sid, sidOk := hashesToSessionId.Load(hash); sidOk {
					toRun = append(toRun, handlerEntry{hash: hash, sessionId: sid, cb: cb})
				}
			}
			return true
		})
	}

	if len(toRun) == 0 {
		return
	}

	// Set the calling flag under the lock, then release before executing handlers.
	lock.Lock()
	callingHandler.Store(true)
	lock.Unlock()

	var wg sync.WaitGroup
	for _, entry := range toRun {
		wg.Add(1)
		go func(he handlerEntry) {
			defer wg.Done()
			he.cb(HandlerData{
				SessionId: he.sessionId,
				Socket:    h.manager.Get(string(he.sessionId)),
				Manager:   h.manager,
			})
		}(entry)
	}
	wg.Wait()

	lock.Lock()
	callingHandler.Store(false)
	lock.Unlock()
}

func (h *MessageHandler) OnClientSideEvent(handlerId string, sessionId session.Id) {
	cb, ok := handlers.Load(handlerId)
	if ok {
		cb(HandlerData{
			SessionId: sessionId,
			Socket:    h.manager.Get(string(sessionId)),
			Manager:   h.manager,
		})
	}
}

func (h *MessageHandler) OnDomElementRemoved(handlerId string) {
	handlers.Delete(handlerId)
}

func (h *MessageHandler) OnSocketDisconnected(event wsutil.SocketEvent) {
	sessionId := session.Id(event.SessionId)
	hashes, ok := sessionIdToHashes.Load(sessionId)
	if ok {
		hashes.Range(func(hash KeyHash, _ bool) bool {
			hashesToSessionId.Delete(hash)
			handlers.Delete(hash)
			return true
		})
		sessionIdToHashes.Delete(sessionId)
	}
	// Clean up session state.
	session.Delete(sessionId)
}
