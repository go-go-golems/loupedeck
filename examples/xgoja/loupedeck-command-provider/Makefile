.DEFAULT_GOAL := smoke

WORKSPACE_ROOT := $(abspath ../../../..)
XGOJA_ROOT := $(WORKSPACE_ROOT)/go-go-goja
BIN := $(CURDIR)/dist/loupedeck-command-provider
XGOJA := cd $(XGOJA_ROOT) && GOWORK=off go run ./cmd/xgoja
PORT := 8791

.PHONY: smoke doctor list build run clean

smoke: doctor list build run

doctor:
	$(XGOJA) doctor -f $(CURDIR)/xgoja.yaml

list:
	$(XGOJA) list-modules -f $(CURDIR)/xgoja.yaml

build:
	$(XGOJA) build -f $(CURDIR)/xgoja.yaml --output $(BIN) --xgoja-replace $(XGOJA_ROOT) --keep-work

run:
	rm -rf $(CURDIR)/dist/web-scene
	( sleep 1; curl -fsS http://127.0.0.1:$(PORT)/state; curl -fsS -X POST http://127.0.0.1:$(PORT)/deal ) &
	$(BIN) loupe web-scene-switcher web-scene-switcher \
		--http-listen 127.0.0.1:$(PORT) \
		--out $(CURDIR)/dist/web-scene \
		--wait-ms 5000
	test -f $(CURDIR)/dist/web-scene/dealt.txt
	test -f $(CURDIR)/dist/web-scene/scene-report.md
	grep -q 'Final scene: dealt' $(CURDIR)/dist/web-scene/scene-report.md

clean:
	rm -rf $(CURDIR)/dist
