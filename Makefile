COMMANDS := $(subst .go,,$(subst cmd,bin,$(wildcard cmd/*.go)))
.PHONY : test-go test-e2e test

default: $(COMMANDS)

bin/%: cmd/%.go
	go build -o $@ $<

clean:
	rm -rf bin

test: test-go test-e2e
	echo "Tests passed! :-)"

test-go:
	go test -v ./lib/...
