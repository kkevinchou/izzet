RELEASE_FOLDER = "build/release"
TAR_FILE = "izzet.tar.gz"
COMPONENTS_PROTO_DIR = "izzet/components/proto"
PLAYERCOMMAND_PROTO_DIR = "izzet/playercommand/proto"
PROTOC_PATH = ~/protoc-21.7-win64/bin/protoc.exe

.PHONY: client
client:
	go run main.go CLIENT

.PHONY: client_no_logs
client_no_logs:
	go run main.go CLIENT --logs=false

# profile fetched from http://localhost:6868/debug/pprof/profile
.PHONY: pprof
pprof:
	go tool pprof -http=localhost:7070 profile

.PHONY: profilecpu
profilecpu:
	curl http://localhost:6868/debug/pprof/profile?seconds=60 -o cpu
	go tool pprof -http=localhost:6969 cpu

.PHONY: profileheap
profileheap:
	curl http://localhost:6868/debug/pprof/heap?seconds=60 -o heap
	go tool pprof -http=localhost:6767 heap

.PHONY: heapsnapshot
heapsnapshot:
	curl -s http://localhost:6868/debug/pprof/heap > heap_snapshot

.PHONY: inuse_repl
inuse_repl:
	go tool pprof --inuse_objects http://localhost:6868/debug/pprof/heap

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	go build -o izzet.exe

.PHONY: server
server:
	go run main.go SERVER

.PHONY: headless
headless:
	go run main.go HEADLESS

.PHONY: clean
clean:
	rm -rf $(RELEASE_FOLDER)
	rm -f $(TAR_FILE)

.PHONY: release
release: clean
	mkdir -p $(RELEASE_FOLDER)
	public_ip=$$(curl -fsS https://api.ipify.org); port=$$(sed -nE 's/.*"server_address"[[:space:]]*:[[:space:]]*"[^"]*:([^"]*)".*/\1/p' config.json); sed -E 's#"server_address"[[:space:]]*:[[:space:]]*"[^"]*"#"server_address": "'"$$public_ip:$$port"'"#' config.json > build/release/config.json
	cp -r shaders $(RELEASE_FOLDER)/
	cp -r _assets $(RELEASE_FOLDER)/
	cp -r .project $(RELEASE_FOLDER)/
	CGO_ENABLED=1 CGO_LDFLAGS="-static" CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build -tags static -ldflags "-s -w" -o $(RELEASE_FOLDER)/izzet.exe
	# tar -zcf $(TAR_FILE) $(RELEASE_FOLDER)
