APPLICATION      := iam-role-annotator
LINUX            := build/${APPLICATION}-linux-amd64
DARWIN           := build/${APPLICATION}-darwin-amd64
DOCKER_USER      ?= ""
DOCKER_PASS      ?= ""
BIN_DIR          := $(GOPATH)/bin
GOMETALINTER     := $(BIN_DIR)/gometalinter
COVER            := $(BIN_DIR)/gocov-xml
JUNITREPORT      := $(BIN_DIR)/go-junit-report
TRAVIS_COMMIT    ?= latest


.PHONY: $(DARWIN)
$(DARWIN):
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -installsuffix cgo -o ${DARWIN} *.go

.PHONY: $(LINUX)
$(LINUX):
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o ${LINUX} *.go

$(GOMETALINTER):
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.16.0

$(COVER):
	go get -u github.com/axw/gocov/gocov
	go get -u github.com/AlekSi/gocov-xml

$(JUNITREPORT):
	go get -u github.com/jstemmer/go-junit-report

.PHONY: lint
lint: $(GOMETALINTER)
	golangci-lint run --disable errcheck

.PHONY: test
test: $(JUNITREPORT)
	go test -v -cover ./... | tee /dev/tty | go-junit-report > junit-report.xml

.PHONY: coverage
coverage: $(COVER) lint
	./coverage

.PHONY: release
release: $(LINUX)
	echo "${DOCKER_PASS}" | docker login -u "${DOCKER_USER}" --password-stdin
	docker build -t "${DOCKER_IMAGE}" "."
	docker tag "${DOCKER_IMAGE}" "${DOCKER_IMAGE}:${TRAVIS_COMMIT}"
	docker push "${DOCKER_IMAGE}"

e2e:
	./e2e_test.sh
