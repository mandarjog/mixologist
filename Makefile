all: test

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
	@ if ! which glide > /dev/null; then \
		echo "error: glide depedency mget not installed not installed" >&2; \
		exit 1; \
	fi

test: build
	go test -v -cpu 1,4 ./mixologist/...

testrace: test
	go test -v -race -cpu 1,4 ./mixologist/...  -coverprofile=coverage.out

clean:
	go clean -i ./...

coverage: build
	./coverage.sh --coveralls

build:
	go build main.go

.PHONY: \
	all \
	build \
	proto \
	test \
	testrace \
	clean \
	coverage
