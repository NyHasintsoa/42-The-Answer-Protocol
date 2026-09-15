FROM golang:1.26.5-bookworm

RUN apt-get update && apt-get install -y \
    gcc \
    libgl1-mesa-dev \
    libxi-dev \
    libxcursor-dev \
    libxrandr-dev \
    libxinerama-dev \
    libwayland-dev \
    libxkbcommon-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=1
RUN go build -o tap-gui main.go