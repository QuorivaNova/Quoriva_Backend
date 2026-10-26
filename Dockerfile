FROM public.ecr.aws/docker/library/golang:1.27.3-bookworm@sha256:972b4110ba0f47539a975c31e52c64765e1ff19f2d74469ae4a595c6676e80a5 AS development
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
RUN go build -trimpath -o /out/api ./cmd/server && go build -trimpath -o /out/migrate ./cmd/migrate

FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab AS runtime
COPY --from=development /out/api /api
COPY --from=development /out/migrate /migrate
EXPOSE 8080
ENTRYPOINT ["/api"]
