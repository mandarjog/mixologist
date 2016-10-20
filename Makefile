TOP := $(shell pwd)
MIXREPO := github.com/cloudendpoints/mixologist


DEV_REPO := gcr.io/$(PROJECT_ID)/mixologist-$(USER)
GLIDE := build/glide.$(shell uname)

all: inst build

inst:
	$(GLIDE) install

check:
ifndef GOPATH
	$(error GOPATH is undefined)
endif
ifeq (,$(findstring $(MIXREPO),$(TOP)))
	$(error project should be built at $(GOPATH)/src/$(MIXREPO))
endif

dep-prep:
	@ if ! which golint > /dev/null; then \
		echo "error: golint not installed" >&2; \
		echo "go get -u github.com/golang/lint/golint" >&2;\
	fi
	@touch dep-prep

test: mixologist-bin
	go test -v -cpu 1,4 ./mixologist/...
	go test -v -race -cpu 1,4 ./mixologist/...
	./script/coverage.sh

clean:
	go clean -i ./...
	rm -f mixologist-bin
	rm -f client

coverage: build
	./script/coverage.sh --html

mixologist-bin: dep-prep main.go mixologist/*.go
	go build -o mixologist-bin main.go

build: check mixologist-bin client main.go mixologist/*.go mixologist/rc/prometheus/*.go
	go vet main.go
	golint main.go
	go vet `go list ./mixologist/...`
	golint mixologist/...

client: clnt/*.go
	go build -o client clnt/check.go

run: mixologist-bin
	./mixologist-bin -v=2 -logtostderr=true


docker: build clean
	docker build -t mixologist .

docker-run:
	docker run -d -p 9092:9092 mixologist -v=1  -logtostderr=true

check-env:
ifndef PROJECT_ID
	$(error PROJECT_ID is undefined)
endif
ifndef NAMESPACE
	$(error NAMESPACE is undefined)
endif

dev-build: check-env build clean
	@echo "Building $(DEV_REPO)"
	docker build -t $(DEV_REPO) .
	gcloud docker push $(DEV_REPO)

dev-deploy: dev-build
	DEMO/deploy.py --namespace $(NAMESPACE) --MIXOLOGIST-IMAGE $(DEV_REPO)

#TODO change this to deployments and replica sets
# then we can use rolling updates
dev-redeploy: dev-build
	kubectl --namespace $(NAMESPACE) scale --replicas=0 rc/mixologist
	kubectl --namespace $(NAMESPACE) scale --replicas=1 rc/mixologist


.PHONY: \
	all \
	proto \
	test \
	clean \
	coverage \
	build \
	run \
	docker \
	docker-run \
	dev-build \
	dev-deploy \
	dev-redeploy \
	check-env \
	check \
	inst
