CC= go build 


dedupe-agent:
	go build -o dedupe-agent -a cmd/dedupe-agent/main.go

dedupe-agent-arm:
	GOARCH=arm64 GOOS=linux go build -o dedupe-agent -a cmd/dedupe-agent/main.go

dedupe-agent-clean:
	rm -f dedupe-agent


clean: dedupe-agent-clean
	rm -f photo-deduplicator

