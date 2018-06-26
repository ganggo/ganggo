# use bash instead of shell we'll
# need the ability to create arrays
SHELL=/bin/bash

version := $(shell echo -n $(VERSION) |cut -d- -f1)
package := github.com/ganggo/ganggo
srcdir := $(GOPATH)/src/$(package)

go := $(shell command -v go 2> /dev/null)
gobin := $(shell command -v go-bindata 2> /dev/null)
npm := $(shell command -v npm 2> /dev/null)
revel := $(shell command -v revel 2> /dev/null)
train := $(shell command -v train 2> /dev/null)

define version_info
Please define a version you want to build e.g. VERSION=v0 make
endef

define install_deps_info
is not available please run 'make install-deps'
endef

define train_info
$(install_deps_info)

And don't forget to add the node modules to your path

e.g. export PATH=$$PATH:$$(pwd)/node_modules/.bin

endef

install: clean install-deps

release: precompile compile u2d-wrapper

install-deps:
ifndef go
	$(error "go is not available: https://golang.org/dl/")
endif
ifndef npm
	$(error "npm is not available please install it first!")
endif
	# Install CSS and Javascript dependencies
	cd $(srcdir) && npm install --prefix .
	# GangGo dependencies
	go get -u github.com/golang/dep/cmd/dep
	cd $(srcdir) && dep ensure
	# XXX this seams to be a bug in revel
	# it cannot find the api module within
	# the vendor directory even though it exists
	go get -d github.com/ganggo/api/...
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
			dir=$(srcdir)/public/assets/vendor/$${obj[0]}; \
			mkdir -p $$dir && echo $$dir; \
			cp -v $${obj[1]} $$dir; \
		done

compile:
ifndef version
	$(error $(version_info))
endif
ifndef revel
	$(error "revel $(install_deps_info)")
endif
	cp $(srcdir)/conf/app.conf.example $(srcdir)/conf/app.conf
	cd $(srcdir) && APP_VERSION=$(version) revel package $(package)

test:
	go tool vet -v -all app/
ifndef revel
	$(error "revel $(install_deps_info)")
endif
	cp $(srcdir)/conf/app.conf.example $(srcdir)/conf/app.conf
	# XXX revel will not print error stacks to console
	# (see https://github.com/revel/cmd/issues/121)
	revel test $(package) ci || { \
		cd $(srcdir) && bash tests/scripts/test_results.sh ;\
		exit 1 ;\
	}

u2d-wrapper:
ifndef version
	$(error $(version_info))
endif
ifndef gobin
	$(error "go-bindata $(install_deps_info)")
endif
	mkdir -p $(srcdir)/tmp && cd $(srcdir)/tmp && \
		tar -x -f "../ganggo.tar.gz" && \
		rm -rf src/$(package)/{vendor,node_modules} && \
		go-bindata -o ../bindata.go ganggo src/...
	cd $(srcdir) && go build \
		-ldflags "-X main.version=$(version)" \
		-o updater updater.go bindata.go
