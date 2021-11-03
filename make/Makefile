
#export flags="-X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.githash=`git describe --long --dirty --abbrev=14`"
export flags="-X 'main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'`' -X 'main.goversion=`go version`'"

.PHONY: build
build:
	#-ldflags $(flags) -x -o
	@echo $(flags)
	go build -ldflags $(flags) -x -o buildversion ./version.go
	chmod +x ./buildversion
