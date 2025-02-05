FROM golang:bullseye
RUN curl -sL https://deb.nodesource.com/setup_22.x | bash - && \
    apt-get install libgtk-3-0 libwebkit2gtk-4.0-dev gcc-aarch64-linux-gnu nodejs jq -y && \
    rm -rf /var/lib/apt/lists/*
RUN go install github.com/wailsapp/wails/v2/cmd/wails@latest