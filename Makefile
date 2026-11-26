# Hollow Ward. Needs Go 1.27 and golangci-lint 2.15+ on PATH.
# gd (the graphics.gd build tool) is pinned by go.mod and runs via `go tool`;
# it fetches Godot 4.7 itself on first use.

GO   ?= go
GD   ?= $(GO) tool gd
GOLANGCI ?= golangci-lint

.PHONY: check vet lint test fuzz smoke shots balance run editor build clean

## check: vet, lint and race tests. Must pass before every commit.
check: vet lint test

vet:
	$(GO) vet ./...

lint:
	$(GOLANGCI) run ./...

test:
	$(GO) test -race ./...

## fuzz: run the hollow generator fuzz test for a while (from M2).
fuzz:
	$(GO) test ./internal/sim/... -run '^$$' -fuzz FuzzGenerateHollow -fuzztime 60s

## smoke: start the game headless and autoplay through the real UI path.
smoke:
	$(GD) run --headless -- --autoplay

## shots: play the smoke script in a real window on a virtual display and save
## a PNG of every screen (Linux, needs xvfb).
SHOTS ?= /tmp/hollow-ward-shots
shots:
	mkdir -p $(SHOTS)
	xvfb-run -a -s "-screen 0 1280x720x24" $(GD) run -- --autoplay --shots=$(SHOTS)

## balance: 1,000 tenures per scripted policy (from M4).
balance:
	$(GO) run ./cmd/balance -n 1000

## run: play in a window.  editor: open the Godot editor on graphics/.
run:
	$(GD) run

editor:
	$(GD)

## build: desktop builds into releases/ (cross-compiles need zig, which gd fetches).
build:
	GOOS=linux   $(GD) build
	GOOS=macos   $(GD) build
	GOOS=windows $(GD) build

clean:
	rm -rf releases graphics/.godot graphics/*.so graphics/*.dll graphics/*.dylib
