package websocket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"github.com/yeecean/oneweb/internal/application"
	"github.com/yeecean/oneweb/internal/domain/operation"
	"github.com/yeecean/oneweb/internal/infrastructure/events"
)

func TestGatewayOperationStream(t *testing.T) {
	bus := events.NewEventBus(16)
	opSvc := application.NewOperationService(bus, nil)
	gw := NewGateway(opSvc)

	op, _ := opSvc.Create("p1", operation.TypeAuth)
	_ = opSvc.Start(op.ID)

	// httptest server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gw.HandleOperationWS(w, r)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 连接 WS（http 升级为 ws）
	wsURL := "ws" + srv.URL[len("http"):] + "/ws/v1/operations/" + op.ID
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")

	// 触发日志事件
	opSvc.AppendLog(op.ID, "hello log", "stdout")

	// 读取帧
	var frame Frame
	if err := wsjson.Read(ctx, conn, &frame); err != nil {
		t.Fatalf("read frame: %v", err)
	}
	if frame.Type != "log" {
		t.Fatalf("frame type = %q, want log", frame.Type)
	}
	if frame.Sequence != 1 {
		t.Fatalf("sequence = %d, want 1", frame.Sequence)
	}
}
