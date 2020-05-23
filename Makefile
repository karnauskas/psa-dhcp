export GO111MODULE=on
COMMANDS := $(subst .go,,$(subst cmd,bin,$(wildcard cmd/*.go)))
PROTOS := $(subst .proto,.pb.go,$(wildcard lib/server/proto/*.proto))
.PHONY : test-go test-e2e test


default: $(COMMANDS) $(PROTOS)

bin/%: cmd/%.go
	go build -o $@ $<

clean:
	rm -rf bin

test: test-go test-e2e
	echo "Tests passed! :-)"

test-go:
	go test -v ./lib/...

rpi:
	env GOOS=linux GOARCH=arm GOARM=5 go build cmd/psa-dhcpc.go


lib/server/proto/%.pb.go: lib/server/proto/%.proto
	protoc --go_out=. $<

lib/oui/oui.go:
	go run lib/oui/gen-liboui.go > lib/oui/oui.go
