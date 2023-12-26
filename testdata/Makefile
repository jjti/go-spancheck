# note: I'm copying https://github.com/ghostiam/protogetter/blob/main/testdata/Makefile
#
# x/tools/go/analysis/analysistest does not support go modules. To work around this issue
# we need to vendor any external modules to `./src`.
#
# Follow https://github.com/golang/go/issues/37054 for more details.

.PHONY: vendor
vendor:
	rm -rf src
	go mod vendor
	mv vendor src
