BIN      := goberzurg
PREFIX   ?= /usr/local
BINDIR   := $(PREFIX)/bin
MANDIR   := $(PREFIX)/share/man/man1

.PHONY: all lib cli install install-cli install-man clean test

all: lib cli

lib:
	go build ./...

cli:
	go build -o $(BIN) ./cmd/goberzurg/

install: install-cli install-man

install-cli: cli
	install -d $(DESTDIR)$(BINDIR)
	install -m 0755 $(BIN) $(DESTDIR)$(BINDIR)/$(BIN)

install-man:
	install -d $(DESTDIR)$(MANDIR)
	install -m 0644 man/goberzurg.1 $(DESTDIR)$(MANDIR)/goberzurg.1

test:
	go test -v -race ./...

clean:
	rm -f $(BIN)
