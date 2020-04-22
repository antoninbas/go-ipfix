GO := go
BINDIR := $(CURDIR)/bin

CGO_LDFLAGS:=-lipfix

export CGO_LDFLAGS
export CGO_CPPFLAGS
export LD_LIBRARY_PATH

all: bin

.PHONY: bin
bin:
	GOBIN=$(BINDIR) $(GO) install github.com/antoninbas/go-ipfix/...

clean:
	rm -rf bin

.PHONY: fmt
fmt:
	$(GO) fmt github.com/antoninbas/go-ipfix/...
