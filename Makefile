RELEASE_FOLDER = "kitorelease"
TAR_FILE = "izzet.tar.gz"
COMPONENTS_PROTO_DIR = "izzet/components/proto"
PLAYERCOMMAND_PROTO_DIR = "izzet/playercommand/proto"
PROTOC_PATH = ~/protoc-21.7-win64/bin/protoc.exe

# profile fetched from http://localhost:6868/debug/pprof/profile
.PHONY: pprof
pprof:
	go tool pprof -http=localhost:7070 profile

.PHONY: profilecpu
profilecpu:
	curl http://localhost:6868/debug/pprof/profile?seconds=20 -o profile
	go tool pprof -http=localhost:6969 profile

.PHONY: profileheap
profileheap:
	curl http://localhost:6868/debug/pprof/heap?seconds=20 -o heap
	go tool pprof -http=localhost:6969 heap

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	go build -o izzet.exe 
