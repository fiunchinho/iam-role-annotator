APPLICATION      := iam-role-annotator
LINUX            := build/${APPLICATION}-linux-amd64
DARWIN           := build/${APPLICATION}-darwin-amd64
DOCKER_USER      ?= ""
DOCKER_PASS      ?= ""
DOCKER_IMAGE     := fiunchinho/${APPLICATION}
BIN_DIR          := $(GOPATH)/bin
GOMETALINTER     := $(BIN_DIR)/gometalinter
COVER            := $(BIN_DIR)/gocov-xml
JUNITREPORT      := $(BIN_DIR)/go-junit-report
DEP              := $(BIN_DIR)/dep
TRAVIS_COMMIT    ?= latest


.PHONY: $(DARWIN)
$(DARWIN): dep
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -installsuffix cgo -o ${DARWIN} ./cmd/...

.PHONY: $(LINUX)
$(LINUX): dep
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o ${LINUX} ./cmd/...

.PHONY: dep
dep: $(DEP)
	dep ensure -v --vendor-only

$(GOMETALINTER):
	go get -u gopkg.in/alecthomas/gometalinter.v1
	gometalinter.v1 --install --update

$(DEP):
	go get -u github.com/golang/dep/...

$(COVER):
	go get -u github.com/axw/gocov/gocov
	go get -u github.com/AlekSi/gocov-xml

$(JUNITREPORT):
	go get -u github.com/jstemmer/go-junit-report

.PHONY: lint
lint: $(GOMETALINTER)
	gometalinter.v1 --vendor --disable-all -E vet -E goconst -E golint -E goimports -E misspell --deadline=50s -j 11 ./... | tee /dev/tty > checkstyle-report.xml

.PHONY: test
test: $(JUNITREPORT)
	go test -v -cover ./... | tee /dev/tty | go-junit-report > junit-report.xml

.PHONY: coverage
coverage: $(COVER) lint
	./coverage

.PHONY: release
release: $(LINUX)
	docker login --username "${DOCKER_USER}" --password "${DOCKER_PASS}"
	docker build -t "${DOCKER_IMAGE}" "."
	docker tag "${DOCKER_IMAGE}" "${DOCKER_IMAGE}:${TRAVIS_COMMIT}"
	docker push "${DOCKER_IMAGE}"
