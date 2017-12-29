# makefile taken from Oragono's one made up by https://github.com/enckse - thanks Sean!
BUILD=./build
WIN=$(BUILD)/win
LINUX=$(BUILD)/linux
OSX=$(BUILD)/osx
ARM6=$(BUILD)/arm
SOURCE=ircdog.go
VERS=XXX

.PHONY: all clean windows osx linux arm6

add-files = mkdir -p $1; \
	cp LICENSE $1; \
	cp ./docs/README $1; \
	mkdir -p $1/docs; \
	cp ./CHANGELOG.md $1/docs/; \
	# cp ./docs/*.md $1/docs/; \
	# cp ./docs/logo* $1/docs/;

all: clean windows osx linux arm6

clean:
	rm -rf $(BUILD)
	mkdir -p $(BUILD)

windows:
	GOOS=windows GOARCH=amd64 go build $(SOURCE)
	$(call add-files,$(WIN))
	mv ircdog.exe $(WIN)
	cd $(WIN) && zip -r ../ircdog-$(VERS)-windows.zip *

osx:
	GOOS=darwin GOARCH=amd64 go build $(SOURCE)
	$(call add-files,$(OSX))
	mv ircdog $(OSX)
	cd $(OSX) && tar -czvf ../ircdog-$(VERS)-osx.tgz *

linux:
	GOOS=linux GOARCH=amd64 go build $(SOURCE)
	$(call add-files,$(LINUX))
	mv ircdog $(LINUX)
	cd $(LINUX) && tar -czvf ../ircdog-$(VERS)-linux.tgz *

arm6:
	GOARM=6 GOARCH=arm go build $(SOURCE)
	$(call add-files,$(ARM6))
	mv ircdog $(ARM6)
	cd $(ARM6) && tar -czvf ../ircdog-$(VERS)-arm.tgz *

deps:
	go get -v -d

test:
	cd lib && go test .
	cd lib && go vet .
