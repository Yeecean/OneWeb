VERSION ?= v0.1.0-dev
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build test fmt vet lint clean web-build web-install build-all install uninstall

# 构建完整二进制（前端已 embed）
build:
	cd web && npm run build
	rm -rf internal/api/rest/static
	cp -r web/dist internal/api/rest/static
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o dist/oneweb ./cmd/oneweb

build-all:
	GOOS=linux GOARCH=amd64 $(MAKE) build
	GOOS=linux GOARCH=arm64 $(MAKE) build

test:
	go test ./... -count=1

fmt:
	gofmt -w ./cmd ./internal

vet:
	go vet ./...

lint: vet

web-install:
	cd web && npm install

web-build:
	cd web && npm run build

# 本地安装到 ~/.local/bin + 安装用户 systemd 单元
install: build
	install -Dm755 dist/oneweb $(HOME)/.local/bin/oneweb
	install -Dm644 packaging/systemd/oneweb.service $(HOME)/.config/systemd/user/oneweb.service
	systemctl --user daemon-reload
	@echo "Installed. Enable with: systemctl --user enable --now oneweb"

uninstall:
	rm -f $(HOME)/.local/bin/oneweb
	rm -f $(HOME)/.config/systemd/user/oneweb.service
	systemctl --user daemon-reload

clean:
	rm -rf dist
	rm -rf web/dist
	rm -rf internal/api/rest/static
