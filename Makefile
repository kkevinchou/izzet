RELEASE_FOLDER = "build/release"
RELEASE_CONFIG_WIDTH = 1024
RELEASE_CONFIG_HEIGHT = 720
RELEASE_CONFIG_FULLSCREEN = true
RELEASE_CONFIG_PROFILE = true
RELEASE_CONFIG_SERVER_PORT = 7878

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

.PHONY: release
release: clean
	mkdir -p $(RELEASE_FOLDER)
	public_ip=$$(curl -fsS https://api.ipify.org) && \
	printf '{\n\t"width": $(RELEASE_CONFIG_WIDTH),\n\t"height": $(RELEASE_CONFIG_HEIGHT),\n\t"fullscreen": $(RELEASE_CONFIG_FULLSCREEN),\n\t"profile": $(RELEASE_CONFIG_PROFILE),\n\t"server_address": "%s:$(RELEASE_CONFIG_SERVER_PORT)"\n}\n' "$$public_ip" > $(RELEASE_FOLDER)/config.json
	cp -r shaders $(RELEASE_FOLDER)/
	cp -r _assets $(RELEASE_FOLDER)/
	cp -r .project $(RELEASE_FOLDER)/
	go build -o $(RELEASE_FOLDER)/izzet.exe
