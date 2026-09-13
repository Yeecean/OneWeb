package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/yeecean/oneweb/internal/api/rest"
	"github.com/yeecean/oneweb/internal/api/websocket"
	"github.com/yeecean/oneweb/internal/application"
	"github.com/yeecean/oneweb/internal/domain/profile"
	"github.com/yeecean/oneweb/internal/domain/runtime"
	"github.com/yeecean/oneweb/internal/infrastructure/events"
	"github.com/yeecean/oneweb/internal/infrastructure/filesystem"
	"github.com/yeecean/oneweb/internal/infrastructure/onedrive"
	"github.com/yeecean/oneweb/internal/infrastructure/runtime/systemd"
)

// version 在构建时通过 -ldflags 注入。
var version = "v0.1.0-dev"

func main() {
	var listen string
	flag.StringVar(&listen, "listen", "127.0.0.1:8080", "HTTP listen address")
	flag.Parse()

	// 兼容 `oneweb serve --listen ...` 子命令形式
	if args := flag.Args(); len(args) > 0 && args[0] == "serve" {
		fs := flag.NewFlagSet("serve", flag.ExitOnError)
		fs.StringVar(&listen, "listen", listen, "HTTP listen address")
		_ = fs.Parse(args[1:])
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Printf("OneWeb %s — starting...\n", version)

	// 1. 基础设施初始化
	cli, err := onedrive.NewCLIExecutor("")
	if err != nil {
		log.Printf("warning: onedrive binary not available: %v", err)
		cli = nil
	}

	// 2. 应用服务初始化
	profileStore := filesystem.DefaultProfileStore()
	profileSvc := application.NewProfileService(profileStore, cli)
	configSvc, err := application.NewConfigService(cli)
	if err != nil {
		log.Fatalf("init config service: %v", err)
	}
	systemdBackend := &systemd.Backend{}
	runtimeSvc := application.NewRuntimeService(map[profile.RuntimeType]runtime.RuntimeBackend{
		profile.RuntimeSystemd: systemdBackend,
	})

	// 事件总线 + 异步操作服务 + 认证服务
	eventBus := events.NewEventBus(64)
	opService := application.NewOperationService(eventBus, cli)
	authSvc := application.NewAuthService(opService, cli)
	wsGateway := &websocket.Gateway{OpService: opService}
	syncSvc := application.NewSyncService()

	// 3. API 路由
	router := rest.NewRouter(&rest.Dependencies{
		Version:      version,
		ProfileSvc:   profileSvc,
		ConfigSvc:    configSvc,
		RuntimeSvc:   runtimeSvc,
		OperationSvc: opService,
		AuthSvc:      authSvc,
		SyncSvc:      syncSvc,
		WSGateway:    wsGateway,
		CLI:          cli,
	})

	server := &http.Server{
		Addr:              listen,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// 4. 启动
	go func() {
		fmt.Printf("OneWeb listening on http://%s\n", listen)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// 5. 优雅关闭
	<-ctx.Done()
	fmt.Println("\nshutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
	fmt.Println("bye")
}
