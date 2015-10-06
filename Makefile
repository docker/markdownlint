
# Adds build information from git repo
#
# as suggested by tatsushid in
# https://github.com/spf13/hugo/issues/540

COMMIT_HASH=`git rev-parse --short HEAD 2>/dev/null`
BUILD_DATE=`date +%FT%T%z`
LDFLAGS=-ldflags "-X github.com/spf13/hugo/hugolib.CommitHash=${COMMIT_HASH} -X github.com/spf13/hugo/hugolib.BuildDate=${BUILD_DATE}"

shell: docker-build
	docker run --rm -it -v $(CURDIR):/go/src/github.com/SvenDowideit/markdownlint markdownlint bash

docker-build:
	rm -f markdownlint.gz
	docker build -t markdownlint .

docker: docker-build
	docker run --name markdownlint-build markdownlint gzip markdownlint
	docker cp markdownlint-build:/go/src/github.com/SvenDowideit/markdownlint/markdownlint.gz .
	docker rm markdownlint-build
	gunzip markdownlint.gz

run:
	./markdownlint .
