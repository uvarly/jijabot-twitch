BINARY_NAME := jijabot
CMD_PATH    := ./cmd/jijabot
BIN_DIR     := bin
LDFLAGS     := -s -w

.PHONY: all build build-linux build-windows clean

all: build

build: build-linux build-windows

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o ${BIN_DIR}/${BINARY_NAME} ${CMD_PATH}

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o ${BIN_DIR}/${BINARY_NAME}.exe ${CMD_PATH} 

clean:
	rm -rf ${BIN_DIR}
