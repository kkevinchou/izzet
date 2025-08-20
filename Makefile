RELEASE_FOLDER = "build/release"
TAR_FILE = "izzet.tar.gz"
COMPONENTS_PROTO_DIR = "izzet/components/proto"
PLAYERCOMMAND_PROTO_DIR = "izzet/playercommand/proto"
PROTOC_PATH = ~/protoc-21.7-win64/bin/protoc.exe

.PHONY: run
run:
	go run main.go

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
	curl http://localhost:6868/debug/pprof/heap?seconds=10 -o heap
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
	CGO_LDFLAGS="-LC:/Users/kkevi/mingw64/x86_64-w64-mingw32/lib -lSDL2" go build -o izzet.exe 

.PHONY: client
client:
	CGO_LDFLAGS="-LC:/Users/kkevi/mingw64/x86_64-w64-mingw32/lib -lSDL2" go run main.go CLIENT

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
	cp config.json $(RELEASE_FOLDER)/
	cp -r shaders $(RELEASE_FOLDER)/
	cp -r _assets $(RELEASE_FOLDER)/
	cp -r .project $(RELEASE_FOLDER)/
	cp izzet_data.json $(RELEASE_FOLDER)/
	CGO_ENABLED=1 CGO_LDFLAGS="-static" CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build -tags static -ldflags "-s -w" -o $(RELEASE_FOLDER)/izzet.exe
	# tar -zcf $(TAR_FILE) $(RELEASE_FOLDER)
