.PHONY: test
test: $(patsubst %.demo,%.svg,$(wildcard ./testdata/*.demo))

.PHONY: clean
clean:
	@rm -f ./testdata/*.cast ./testdata/*.svg ./testdata/*.mp4

.PRECIOUS: %.cast
%.cast: %.demo
	@WORK_DIR=$(shell dirname $<) \
	ROOT_DIR=$(shell pwd) \
	go run ./cmd/democtl rec -i "$<" -o "$@"

.PRECIOUS: %.svg
%.svg: %.cast
	@go run ./cmd/democtl svg -i "$<" -o "$@"

%.mp4: %.cast
	@go run ./cmd/democtl mp4 -i "$<" -o "$@"

%.gif: %.cast
	@go run ./cmd/democtl gif -i "$<" -o "$@"
