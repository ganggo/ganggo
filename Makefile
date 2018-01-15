version := $(shell echo -n $(VERSION) |cut -d- -f1)
ifndef version
$(error "Please define a version you want to build e.g. VERSION=v0 make")
endif

package := gopkg.in/ganggo/ganggo.$(version)
srcdir := $(GOPATH)/src/$(package)

go := $(shell command -v go 2> /dev/null)
gobin := $(shell command -v go-bindata 2> /dev/null)
npm := $(shell command -v npm 2> /dev/null)
revel := $(shell command -v revel 2> /dev/null)
train := $(shell command -v train 2> /dev/null)

define install_deps_info
is not available please run 'make install-deps'
endef

define train_info
$(install_deps_info)

And don't forget to add the node modules to your path

e.g. export PATH=$$PATH:$$(pwd)/node_modules/.bin

endef

all: clean precompile compile

release: all u2d-wrapper

install-deps:
ifndef go
	$(error "go is not available: https://golang.org/dl/")
endif
ifndef npm
	$(error "npm is not available please install it first!")
endif
	# Install CSS and Javascript dependencies
	npm install
	# GangGo dependencies
	go get -d \
		./... \
		gopkg.in/ganggo/api.$(version)/... \
		gopkg.in/ganggo/federation.$(version)/...

	## CLI for train asset library / revel webframework
	go get -d \
		github.com/shaoshing/train \
		github.com/revel/cmd/...
	go build -o $(GOPATH)/bin/train github.com/shaoshing/train/cmd
	go build -o $(GOPATH)/bin/revel github.com/revel/cmd/revel
	# Embedding binary data e.g. assets
	go get -u github.com/jteeuwen/go-bindata/...

clean:
	rm -r app/assets/vendor/* \
		tmp node_modules || true
	rm bindata.go updater *.tar.gz || true

precompile:
ifndef train
	$(error "train $(train_info)")
endif
	cd $(srcdir) && train -out public -source app/assets
	# Append vendor files to manifest
	sed -n '/^.*:node_modules.*/p' .include_vendor \
		| while read line; do \
			obj=($${line/:/ }); \
			dir=public/assets/vendor/$${obj[0]}; \
			mkdir -p $$dir && echo $$dir; \
			cp -v $${obj[1]} $$dir; \
		done

compile:
ifndef revel
	$(error "revel $(install_deps_info)")
endif
	cp $(srcdir)/conf/app.conf.example $(srcdir)/conf/app.conf
	revel package $(package)

u2d-wrapper:
ifndef gobin
	$(error "go-bindata $(install_deps_info)")
endif
	mkdir tmp && tar -x -f "ganggo.$(version).tar.gz" -C tmp
	cd tmp && mv ganggo.$(version)* ganggo && \
		go-bindata -o ../bindata.go ganggo src/...
	go build -ldflags "-X main.version=$(version)" -o updater updater.go bindata.go
