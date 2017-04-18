.PHONY: all build test

all: clean build test

build:
	go build

test:
	bash -c "./script/test.sh"

clean:
	rm -f ./dcos-checks
