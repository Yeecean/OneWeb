// Package websocket 实现 WebSocket 事件扇出网关。
package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"github.com/yeecean/oneweb/internal/application"
	"github.com/yeecean/oneweb/internal/domain"
)

// Frame 是 WebSocket 协议帧。
type Frame struct {
	Sequence  int64       `json:"sequence"`
	Timestamp string      `json:"timestamp"`
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
}

// Gateway 管理 WebSocket 连接的生命周期与事件推送。
type Gateway struct {
	OpService *application.OperationService

	mu  sync.Mutex
	seq int64
}

// NewGateway 创建 WebSocket 网关。
func NewGateway(opService *application.OperationService) *Gateway {
	return &Gateway{OpService: opService}
}

// HandleOperationWS 处理 /ws/v1/operations/{id}。
func (g *Gateway) HandleOperationWS(w http.ResponseWriter, r *http.Request) {
	opID := r.URL.Path[len("/ws/v1/operations/"):]
	if opID == "" {
		http.Error(w, "missing operation id", http.StatusBadRequest)
		return
	}
	op, err := g.OpService.Get(opID)
	if err != nil {
		http.Error(w, "operation not found", http.StatusNotFound)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // 本地/内网部署；生产建议反代 TLS
	})
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusInternalError, "closed")

	ctx := r.Context()
	sub := g.OpService.EventBus.Subscribe(domain.EvtOpLog)
	defer g.OpService.EventBus.Unsubscribe(sub)
	subAll := g.OpService.EventBus.Subscribe(domain.EvtOpStarted)
	defer g.OpService.EventBus.Unsubscribe(subAll)
	subNeed := g.OpService.EventBus.Subscribe(domain.EvtOpNeedInput)
	defer g.OpService.EventBus.Unsubscribe(subNeed)
	subDone := g.OpService.EventBus.Subscribe(domain.EvtOpCompleted)
	defer g.OpService.EventBus.Unsubscribe(subDone)
	subFail := g.OpService.EventBus.Subscribe(domain.EvtOpFailed)
	defer g.OpService.EventBus.Unsubscribe(subFail)

	// 回放操作已有日志
	for _, line := range op.Logs {
		g.send(ctx, conn, "log", map[string]interface{}{
			"stream": "stdout",
			"level":  "info",
			"raw":    line,
		})
	}

	// 事件泵
	for {
		select {
		case <-ctx.Done():
			return
		case ev := <-sub:
			if ev.ProfileID == op.ProfileID {
				g.send(ctx, conn, "log", ev.Payload)
			}
		case ev := <-subNeed:
			if ev.ProfileID == op.ProfileID {
				g.send(ctx, conn, "waiting_input", ev.Payload)
			}
		case ev := <-subDone:
			if ev.ProfileID == op.ProfileID {
				g.send(ctx, conn, "completed", ev.Payload)
				return
			}
		case ev := <-subFail:
			if ev.ProfileID == op.ProfileID {
				g.send(ctx, conn, "failed", ev.Payload)
				return
			}
		}
	}
}

func (g *Gateway) send(ctx context.Context, conn *websocket.Conn, typ string, data interface{}) error {
	g.mu.Lock()
	g.seq++
	seq := g.seq
	g.mu.Unlock()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return wsjson.Write(ctx, conn, Frame{
		Sequence:  seq,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Type:      typ,
		Data:      data,
	})
}

// jsonEncode 预留：用于调试。
func jsonEncode(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
