.PHONY: build

BINARY_NAME := yaml2csv
PLATFORMS := windows linux darwin
os = $(word 1, $@)

build:
	go build -o bin/$(BINARY_NAME) *.go

clean:
	rm bin/$(BINARY_NAME)*

.PHONY: $(PLATFORMS)
$(PLATFORMS):
	GOOS=$(os) GOARCH=amd64 go build -o bin/$(BINARY_NAME)-$(os)-amd64 *.go

.PHONY: release
release: $(PLATFORMS)

