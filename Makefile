BIN      := goberzurg
PREFIX   ?= /usr/local
BINDIR   := $(PREFIX)/bin
MAN1DIR  := $(PREFIX)/share/man/man1
MAN3DIR  := $(PREFIX)/share/man/man3

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
	install -d $(DESTDIR)$(MAN1DIR) $(DESTDIR)$(MAN3DIR)
	install -m 0644 man/goberzurg.1 $(DESTDIR)$(MAN1DIR)/goberzurg.1
	install -m 0644 man/goberzurg.3 $(DESTDIR)$(MAN3DIR)/goberzurg.3

test:
	go test -v -race ./...

clean:
	rm -f $(BIN)
