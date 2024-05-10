.PHONY: test
test: $(patsubst %.demo,%.svg,$(wildcard ./testdata/*.demo))

.PHONY: clean
clean:
	@rm -f ./testdata/*.cast ./testdata/*.svg ./testdata/*.mp4

.PRECIOUS: %.cast
%.cast: %.demo
	@WORK_DIR=$(shell dirname $<) \
	ROOT_DIR=$(shell pwd) \
	./democtl.sh "$<" "$@" \
		--ps1='\033[1;96m~/wzshiming/democtl\033[1;94m$$\033[0m ' \
		--shell bash \
		--env WORK_DIR \
		--env ROOT_DIR

.PRECIOUS: %.svg
%.svg: %.cast
	@./democtl.sh "$<" "$@" \
		--term xresources \
	  	--profile ./.xresources

%.mp4: %.cast
	@./democtl.sh "$<" "$@"
