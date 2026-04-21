FROM golang:1.21.3-bookworm

COPY go.mod go.sum* /
RUN cd / && go mod download

ENV APP_DIR=/go/src/app
WORKDIR ${APP_DIR}
COPY . ${APP_DIR}
RUN cd ${APP_DIR}/cmd/main-yugioh-card-editor && go build

CMD ["bash", "-c", "${APP_DIR}/cmd/main-yugioh-card-editor/main-yugioh-card-editor"]
