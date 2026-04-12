package wsutil

import (
	"fmt"
	"github.com/franchb/htmgo/extensions/websocket/opts"
	"github.com/franchb/htmgo/framework/h"
	"github.com/franchb/htmgo/framework/service"
	"github.com/puzpuzpuz/xsync/v3"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type EventType string
type WriterChan chan string
type DoneChan chan bool

const (
	ConnectedEvent    EventType = "connected"
	DisconnectedEvent EventType = "disconnected"
	MessageEvent      EventType = "message"
)

type SocketEvent struct {
	SessionId string
	RoomId    string
	Type      EventType
	Payload   map[string]any
}

type CloseEvent struct {
	Code   int
	Reason string
}

type SocketConnection struct {
	Id     string
	RoomId string
	Done   DoneChan
	Writer WriterChan
}

type ManagerMetrics struct {
	RunningGoroutines   int32
	TotalSockets        int
	TotalRooms          int
	TotalListeners      int
	SocketsPerRoomCount map[string]int
	SocketsPerRoom      map[string][]string
	TotalMessages       int64
	MessagesPerSecond   int
	SecondsElapsed      int
}

type SocketManager struct {
	sockets           *xsync.MapOf[string, *xsync.MapOf[string, SocketConnection]]
	idToRoom          *xsync.MapOf[string, string]
	listeners         []chan SocketEvent
	goroutinesRunning atomic.Int32
	opts              *opts.ExtensionOpts
	lock              sync.Mutex
	totalMessages     atomic.Int64
	messagesPerSecond int
	secondsElapsed    int
	closeOnce         sync.Once
	done              chan struct{}
}

func (manager *SocketManager) StartMetrics() {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-manager.done:
				return
			case <-ticker.C:
				manager.lock.Lock()
				manager.secondsElapsed++
				totalMessages := manager.totalMessages.Load()
				manager.messagesPerSecond = int(float64(totalMessages) / float64(manager.secondsElapsed))
				manager.lock.Unlock()
			}
		}
	}()
}

// Close stops the background metrics goroutine. Call this during graceful
// shutdown to avoid leaking the goroutine started by StartMetrics.
func (manager *SocketManager) Close() {
	manager.closeOnce.Do(func() {
		close(manager.done)
	})
}

func (manager *SocketManager) Metrics() ManagerMetrics {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	count := manager.goroutinesRunning.Load()
	metrics := ManagerMetrics{
		RunningGoroutines:   count,
		TotalSockets:        0,
		TotalRooms:          0,
		TotalListeners:      len(manager.listeners),
		SocketsPerRoom:      make(map[string][]string),
		SocketsPerRoomCount: make(map[string]int),
		TotalMessages:       manager.totalMessages.Load(),
		MessagesPerSecond:   manager.messagesPerSecond,
		SecondsElapsed:      manager.secondsElapsed,
	}

	roomMap := make(map[string]int)

	manager.idToRoom.Range(func(socketId string, roomId string) bool {
		roomMap[roomId]++
		return true
	})

	metrics.TotalRooms = len(roomMap)

	manager.sockets.Range(func(roomId string, sockets *xsync.MapOf[string, SocketConnection]) bool {
		metrics.SocketsPerRoomCount[roomId] = sockets.Size()
		sockets.Range(func(socketId string, conn SocketConnection) bool {
			if metrics.SocketsPerRoom[roomId] == nil {
				metrics.SocketsPerRoom[roomId] = []string{}
			}
			metrics.SocketsPerRoom[roomId] = append(metrics.SocketsPerRoom[roomId], socketId)
			metrics.TotalSockets++
			return true
		})
		return true
	})

	return metrics
}

func SocketManagerFromCtx(ctx *h.RequestContext) *SocketManager {
	locator := ctx.ServiceLocator()
	return service.Get[SocketManager](locator)
}

func NewSocketManager(opts *opts.ExtensionOpts) *SocketManager {
	return &SocketManager{
		sockets:           xsync.NewMapOf[string, *xsync.MapOf[string, SocketConnection]](),
		idToRoom:          xsync.NewMapOf[string, string](),
		opts:              opts,
		goroutinesRunning: atomic.Int32{},
		done:              make(chan struct{}),
	}
}

func (manager *SocketManager) ForEachSocket(roomId string, cb func(conn SocketConnection)) {
	sockets, ok := manager.sockets.Load(roomId)
	if !ok {
		return
	}
	sockets.Range(func(id string, conn SocketConnection) bool {
		cb(conn)
		return true
	})
}

func (manager *SocketManager) RunIntervalWithSocket(socketId string, interval time.Duration, cb func() bool) {
	socketIdSlog := slog.String("socketId", socketId)
	slog.Debug("ws-extension: starting every loop", socketIdSlog, slog.Duration("duration", interval))

	go func() {
		manager.goroutinesRunning.Add(1)
		defer manager.goroutinesRunning.Add(-1)
		tries := 0
		for {
			socket := manager.Get(socketId)
			// This can run before the socket is established, lets try a few times and kill it if socket isn't connected after a bit.
			if socket == nil {
				if tries > 200 {
					slog.Debug("ws-extension: socket disconnected, killing goroutine", socketIdSlog)
					return
				} else {
					time.Sleep(time.Millisecond * 15)
					tries++
					slog.Debug("ws-extension: socket not connected yet, trying again", socketIdSlog, slog.Int("attempt", tries))
					continue
				}
			}
			success := cb()
			if !success {
				return
			}
			time.Sleep(interval)
		}
	}()
}

func (manager *SocketManager) Listen(listener chan SocketEvent) {
	if manager.listeners == nil {
		manager.listeners = make([]chan SocketEvent, 0)
	}
	if listener != nil {
		manager.listeners = append(manager.listeners, listener)
	}
}

// dispatch sends an event to all listeners using non-blocking sends.
// This is intentional: a slow or full listener must not block broadcasts to
// other listeners or to WebSocket writers. Callers that need reliable delivery
// should use a sufficiently buffered channel via Listen().
func (manager *SocketManager) dispatch(event SocketEvent) {
	for _, listener := range manager.listeners {
		select {
		case listener <- event:
		default:
			fmt.Printf("ws-extension: listener channel full, dropping event: %s\n", event.Type)
		}
	}
}

func (manager *SocketManager) OnMessage(id string, message map[string]any) {
	socket := manager.Get(id)
	if socket == nil {
		return
	}

	manager.totalMessages.Add(1)
	manager.dispatch(SocketEvent{
		SessionId: id,
		Type:      MessageEvent,
		Payload:   message,
		RoomId:    socket.RoomId,
	})
}

func (manager *SocketManager) Add(roomId string, id string, writer WriterChan, done DoneChan) {
	manager.idToRoom.Store(id, roomId)

	sockets, _ := manager.sockets.LoadOrCompute(roomId, func() *xsync.MapOf[string, SocketConnection] {
		return xsync.NewMapOf[string, SocketConnection]()
	})

	sockets.Store(id, SocketConnection{
		Id:     id,
		Writer: writer,
		RoomId: roomId,
		Done:   done,
	})

	s, ok := sockets.Load(id)
	if !ok {
		return
	}

	manager.dispatch(SocketEvent{
		SessionId: s.Id,
		Type:      ConnectedEvent,
		RoomId:    s.RoomId,
		Payload:   map[string]any{},
	})
}

func (manager *SocketManager) OnClose(id string) {
	socket := manager.Get(id)
	if socket == nil {
		return
	}
	slog.Debug("ws-extension: removing socket from manager", slog.String("socketId", id))
	manager.dispatch(SocketEvent{
		SessionId: id,
		Type:      DisconnectedEvent,
		RoomId:    socket.RoomId,
		Payload:   map[string]any{},
	})
	roomId, ok := manager.idToRoom.Load(id)
	if !ok {
		return
	}
	sockets, ok := manager.sockets.Load(roomId)
	if !ok {
		return
	}
	sockets.Delete(id)
	manager.idToRoom.Delete(id)
	slog.Debug("ws-extension: removed socket from manager", slog.String("socketId", id))

}

func (manager *SocketManager) CloseWithMessage(id string, message string) {
	conn := manager.Get(id)
	if conn != nil {
		defer manager.OnClose(id)
		manager.writeText(*conn, message)
		conn.Done <- true
	}
}

func (manager *SocketManager) Disconnect(id string) {
	conn := manager.Get(id)
	if conn != nil {
		manager.OnClose(id)
		conn.Done <- true
	}
}

func (manager *SocketManager) Get(id string) *SocketConnection {
	roomId, ok := manager.idToRoom.Load(id)
	if !ok {
		return nil
	}
	sockets, ok := manager.sockets.Load(roomId)
	if !ok {
		return nil
	}
	conn, ok := sockets.Load(id)
	return &conn
}

func (manager *SocketManager) Ping(id string) bool {
	conn := manager.Get(id)
	if conn != nil {
		return manager.writeText(*conn, "ping")
	}
	return false
}

func (manager *SocketManager) writeCloseRaw(writer WriterChan, message string) {
	manager.writeTextRaw(writer, message)
}

func (manager *SocketManager) writeTextRaw(writer WriterChan, message string) {
	// Fast path: try non-blocking send first to avoid timer allocation.
	select {
	case writer <- message:
		return
	default:
	}
	// Slow path: wait with a reusable timer.
	t := time.NewTimer(3 * time.Second)
	defer t.Stop()
	select {
	case writer <- message:
	case <-t.C:
		fmt.Printf("could not send %s to channel after 3s\n", message)
	}
}

func (manager *SocketManager) writeText(socket SocketConnection, message string) bool {
	if socket.Writer == nil {
		return false
	}
	manager.writeTextRaw(socket.Writer, message)
	return true
}

func (manager *SocketManager) BroadcastText(roomId string, message string, predicate func(conn SocketConnection) bool) {
	sockets, ok := manager.sockets.Load(roomId)

	if !ok {
		return
	}

	var wg sync.WaitGroup
	sockets.Range(func(id string, conn SocketConnection) bool {
		if predicate(conn) {
			wg.Add(1)
			go func(c SocketConnection) {
				defer wg.Done()
				manager.writeText(c, message)
			}(conn)
		}
		return true
	})
	wg.Wait()
}

func (manager *SocketManager) SendHtml(id string, message string) bool {
	conn := manager.Get(id)
	minified := strings.ReplaceAll(message, "\n", "")
	minified = strings.ReplaceAll(minified, "\t", "")
	minified = strings.TrimSpace(minified)
	if conn != nil {
		return manager.writeText(*conn, minified)
	}
	return false
}

func (manager *SocketManager) SendText(id string, message string) bool {
	conn := manager.Get(id)
	if conn != nil {
		return manager.writeText(*conn, message)
	}
	return false
}
