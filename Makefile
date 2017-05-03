srcdir := src/gopkg.in/ganggo/ganggo.v0

mode := $(shell echo -n $(MODE) 2> /dev/null)

go := $(shell command -v go 2> /dev/null)
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

release: all package-wrapper

install-deps:
ifndef go
	$(error "go is not available: https://golang.org/dl/")
endif
ifndef npm
	$(error "npm is not available please install it first!")
endif
	# Asset pipeline
	npm install node-sass
	npm install coffee-script
	go get github.com/shaoshing/train
	go build -o $(GOPATH)/bin/train github.com/shaoshing/train/cmd

	# Revel web framework
	go get \
		github.com/revel/revel \
		github.com/ganggo/cmd/revel

	# Document Parser
	go get github.com/PuerkitoBio/goquery

	# Bcrypt password hashing
	go get golang.org/x/crypto/bcrypt

	# ORM
	go get github.com/jinzhu/gorm

	# GangGo
	go get -u \
		gopkg.in/ganggo/federation.v0 \
		gopkg.in/ganggo/api.v0 || true;

clean:
	rm -r test-results || true ; \
	rm -r routes || true ; \
	rm log/* || true ; \
	rm *.tar.gz || true ; \
	rm -r tmp || true ; \
	rm -r node_modules || true ;

precompile:
ifndef train
	$(error "train $(train_info)")
endif
	train -out public -source app/assets

compile:
ifndef revel
	$(error "revel $(install_deps_info)")
endif
	revel package gopkg.in/ganggo/ganggo.v0 $(mode)

package-wrapper:
	tarball="ganggo.v0-$$GOOS.$$GOARCH.tar.gz" ; \
	if [[ "$$GOOS" == "" && "$$GOARCH" == "" ]] ; \
		then tarball="ganggo.v0.tar.gz" ; \
	fi ; \
	mkdir -p tmp ; \
	tar -x -f $$tarball -C tmp ; \
	cd tmp ; \
	cp -v $(srcdir)/conf/app.conf.example $(srcdir)/conf/app.conf ; \
	ln -s $(srcdir)/conf/app.conf ; \
	ln -s $(srcdir)/public ; \
	rm -r $(srcdir)/node_modules ; \
	cd - ; \
	tar -c -z -C tmp -f $$tarball . ;
