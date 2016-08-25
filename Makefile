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
	go vet main.go 
	golint main.go 
	go vet mixologist/*.go
	golint mixologist/*.go
	go build -o mixologist-bin main.go

build: mixologist-bin

run: mixologist-bin
	./mixologist-bin -v=1 -logtostderr=true

.PHONY: \
	all \
	proto \
	test \
	testrace \
	clean \
	coverage \
	build \
	run
