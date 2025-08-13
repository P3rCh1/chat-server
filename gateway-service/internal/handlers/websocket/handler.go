package websocket

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/P3rCh1/chat-server/gateway-service/internal/config"
	"github.com/P3rCh1/chat-server/gateway-service/internal/gateway"
	sessionpb "github.com/P3rCh1/chat-server/gateway-service/shared/proto/gen/go/session"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	handlersInRoom = make(map[int64]map[*connectionHandler]struct{})
	mu             sync.RWMutex
)

type connectionHandler struct {
	uid        int64
	roomID     int64
	conn       *websocket.Conn
	ws         *WS
	writeMutex sync.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
	closeDone  chan struct{}
}

func (h *connectionHandler) setOptions(cfg *config.Websocket) {
	h.conn.SetReadLimit(int64(cfg.MsgMaxSize))
	h.conn.SetReadDeadline(time.Now().Add(cfg.PongWait))
	h.conn.SetPongHandler(func(string) error {
		h.conn.SetReadDeadline(time.Now().Add(cfg.PongWait))
		return nil
	})
}

func newConn(conn *websocket.Conn, uid int64, ws *WS) *connectionHandler {
	h := &connectionHandler{
		uid:       uid,
		roomID:    -1,
		conn:      conn,
		ws:        ws,
		closeDone: make(chan struct{}),
	}
	h.ctx, h.cancel = context.WithCancel(context.Background())
	return h
}

func Connector(
	cfg *config.Config,
	services *gateway.Services,
) http.HandlerFunc {
	const op = "websocket.Connector"
	upgrader := newUpgrader(cfg.Websocket)
	ws := newWS(&cfg.Websocket, services)
	StartKafkaWorkers(ws, cfg.Kafka.WorkerCount)
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		uid, err := services.Session.Verify(r.Context(), &sessionpb.VerifyRequest{Token: token})
		if err != nil {
			if status, ok := status.FromError(err); ok && status.Code() == codes.Internal {
				services.Log.Error(op, "error", status.Message())
			}
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			services.Log.Error(op, "error", err)
			return
		}
		h := newConn(conn, uid.UID, ws)
		h.setOptions(&cfg.Websocket)
		h.enter(0)
		go h.reader()
		go h.pinger()
	}
}

func Shutdown(ctx context.Context) error {
	var grace = true
	stopWorkers()
	var wg sync.WaitGroup
	mu.RLock()
	for _, handlersInRoom := range handlersInRoom {
		wg.Add(len(handlersInRoom))
		for h := range handlersInRoom {
			go func(h *connectionHandler) {
				defer wg.Done()
				h.cancel()
				select {
				case <-ctx.Done():
					grace = false
					return
				case <-h.closeDone:
					return
				}
			}(h)
		}
	}
	mu.RUnlock()
	wg.Wait()
	if !grace {
		return errors.New("shutdown timer end before close websocket")
	}
	return nil
}
