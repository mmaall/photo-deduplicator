CC= go build 


dedupe-agent:
	GOARCH=amd64 GOOS="linux" go build -o dedupe-agent -a cmd/dedupe-agent/main.go

dedupe-agent-arm:
	GOARCH=arm64 GOOS=linux go build -o dedupe-agent -a cmd/dedupe-agent/main.go

dedupe-agent-clean:
	rm -f dedupe-agent

performance-test:
	GOARCH="arm64" GOOS="linux" && go build -o dedupe-agent-amd64 -a cmd/dedupe-agent/main.go
	GOARCH="amd64" GOOS="linux" && go build -o dedupe-agent-aarch -a cmd/dedupe-agent/main.go

	mv dedupe-agent-* test/performance/

performance-test-clean:
	rm -f test/performance/dedupe-agent-*

clean: dedupe-agent-clean performance-test-clean
	rm -f photo-deduplicator

