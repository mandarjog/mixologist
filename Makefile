DEV_REPO := gcr.io/$(PROJECT_ID)/mixologist-$(USER)

all: build

proto:
	@ if ! which protoc > /dev/null; then \
		echo "error: protoc not installed" >&2; \
		exit 1; \
	fi
	go get -u -v github.com/golang/protobuf/protoc-gen-go
	# use $$dir as the root for all proto files in the same directory
	for dir in $$(git ls-files '*.proto' | xargs -n1 dirname | uniq); do \
		protoc -I $$dir --go_out=plugins=grpc:$$dir $$dir/*.proto; \
	done

dep-prep:
	@ if ! which golint > /dev/null; then \
		echo "error: golint not installed" >&2; \
		echo "go get -u github.com/golang/lint/golint" >&2;\
		exit 1;\
	fi
	@ if ! which glide > /dev/null; then \
		echo "error: glide depedency not installed not installed" >&2; \
		echo "error: https://github.com/Masterminds/glide/releases/tag/v0.11.1 " >&2; \
		exit 1; \
	fi
	@touch dep-prep

test: mixologist-bin
	go test -v -cpu 1,4 ./mixologist/...

testrace: test
	go test -v -race -cpu 1,4 ./mixologist/...  -coverprofile=coverage.out

clean:
	go clean -i ./... 
	rm -f mixologist-bin

coverage: build
	./coverage.sh --coveralls

mixologist-bin: dep-prep main.go mixologist/*.go
	go build -o mixologist-bin main.go

build: mixologist-bin main.go mixologist/*.go mixologist/rc/prometheus/*.go 
	go vet main.go 
	golint main.go 
	go vet mixologist/*.go
	go vet mixologist/rc/prometheus/*.go
	golint mixologist/...


run: mixologist-bin
	./mixologist-bin -v=1 -logtostderr=true


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
	testrace \
	clean \
	coverage \
	build \
	run \
	docker \
	docker-run \
	dev-build \
	dev-deploy \
	dev-redeploy \
	check-env
