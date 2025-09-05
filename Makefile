VERSION = $(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.version=$(VERSION)"
OSARCH=$(shell go env GOHOSTOS)-$(shell go env GOHOSTARCH)

NANOAXM=\
	nanoaxm-darwin-amd64 \
	nanoaxm-darwin-arm64 \
	nanoaxm-linux-amd64 \
	nanoaxm-linux-arm64 \
	nanoaxm-linux-arm \
	nanoaxm-windows-amd64.exe

SUPPLEMENTAL=tools/*.sh

my: nanoaxm-$(OSARCH)

$(NANOAXM): cmd/nanoaxm
	GOOS=$(word 2,$(subst -, ,$@)) GOARCH=$(word 3,$(subst -, ,$(subst .exe,,$@))) go build $(LDFLAGS) -o $@ ./$<

nanoaxm-%-$(VERSION).zip: nanoaxm-%.exe $(SUPPLEMENTAL)
	rm -rf $@ $(subst .zip,,$@)
	mkdir $(subst .zip,,$@)
	echo $^ | xargs -n 1 | cpio -pdmu $(subst .zip,,$@)
	zip -r $@ $(subst .zip,,$@)
	rm -rf $(subst .zip,,$@)

nanoaxm-%-$(VERSION).zip: nanoaxm-% $(SUPPLEMENTAL)
	rm -rf $@ $(subst .zip,,$@)
	mkdir $(subst .zip,,$@)
	echo $^ | xargs -n 1 | cpio -pdmu $(subst .zip,,$@)
	zip -r $@ $(subst .zip,,$@)
	rm -rf $(subst .zip,,$@)

clean:
	rm -rf nanoaxm-*

release: $(foreach bin,$(NANOAXM),$(subst .exe,,$(bin))-$(VERSION).zip)

test:
	go test -v -cover -race ./...

.PHONY: my $(NANOAXM) clean release test
