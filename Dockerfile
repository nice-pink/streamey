ARG TARGET

FROM cgr.dev/chainguard/go:latest-dev AS builder
ARG TARGET=cleaner

# Info
LABEL org.opencontainers.image.authors="r@nice.pink"
LABEL org.opencontainers.image.source="https://github.com/nice-pink/streamey/blob/main/Dockerfile"

WORKDIR /app

# get go module ready
COPY go.mod go.sum ./
RUN go mod download

# copy module code
COPY . .

# build all
RUN ./build

FROM cgr.dev/chainguard/glibc-dynamic:latest AS runner

# Info
LABEL org.opencontainers.image.authors="r@nice.pink"
LABEL org.opencontainers.image.source="https://github.com/nice-pink/streamey/blob/main/Dockerfile"

WORKDIR /app

# copy executable
COPY --from=builder /app/bin/* /app/
# ENTRYPOINT [ "/app/streamey" ]

FROM cgr.dev/chainguard/glibc-dynamic:latest AS streamey

# Info
LABEL org.opencontainers.image.authors="r@nice.pink"
LABEL org.opencontainers.image.source="https://github.com/nice-pink/streamey/blob/main/Dockerfile"

WORKDIR /app

# copy executable
COPY --from=builder /app/bin/* ./
COPY --from=builder /app/audios/same_no_tag.mp3 ./same_no_tag.mp3

ENTRYPOINT [ "/app/streamey" ]
