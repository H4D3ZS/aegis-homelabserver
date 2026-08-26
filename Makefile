.PHONY: all build frontend backend install clean

all: build

build: frontend backend

frontend:
	@echo "==> Building Next.js static UI export..."
	cd frontend && npm install && npm run build
	mkdir -p backend/cmd/sentinel/frontend
	cp -r frontend/out backend/cmd/sentinel/frontend/

backend:
	@echo "==> Compiling native single Go binary..."
	mkdir -p bin
	cd backend && go mod tidy && CGO_ENABLED=0 go build -ldflags="-s -w" -o ../bin/pi-sentinel ./cmd/sentinel

install: build
	@echo "==> Running automated bare-metal installer..."
	sudo bash install.sh

clean:
	rm -rf bin frontend/out frontend/.next backend/cmd/sentinel/frontend
