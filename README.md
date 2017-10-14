# GangGo - Diaspora in GoLang

This is the application server repository. Which is still
in pre-alpha state so please don't use it for production!

## Dependencies

For generating the required assets GangGo needs npm for `sass` and `coffee-script`
which works in combination with `github.com/shaoshing/train`

Following golang dependecnies are required for:
 - the web framework `github.com/revel/revel`
 - parsing documents e.g. hcard `github.com/PuerkitoBio/goquery`
 - using password hashing `golang.org/x/crypto/bcrypt`
 - support multiple databases `github.com/jinzhu/gorm`
 - captcha's for registrations `github.com/dchest/captcha`

All this can be installed via:

    make install-deps

## Configuration

Before you can start the application server  
you have to point revel to the right settings.

Copy the example file and adjust the `[ganggo]` section:

    cp conf/app.conf.example conf/app.conf

# Database setup

Please check [the wiki page](https://github.com/ganggo/ganggo/wiki/Database-setup) to setup your database.

## Precompile and Build

Make sure your `node_modules/.bin` is in your `$PATH` variable e.g.:

    export PATH=$PATH:$(pwd)/node_modules/.bin

Then run

    make

## Development

If you don't want to compile the whole application everytime
you change something you can run it directly via

    revel run gopkg.in/ganggo/ganggo.v0

Revel is able to watch changes and recompile if necessary!

**This does not apply to asset files!**  
You should re-compile everytime something changes in `app/assets`.

### Assets

To recompile your assets in your development environment run

    make precompile

### Cross Compile

You can cross compile for multiple systems and architectures.
Simply run e.g.:

    GOOS=linux GOARCH=amd64 make compile

Supported Systems
 - android
 - darwin
 - dragonfly
 - freebsd
 - linux
 - nacl
 - netbsd
 - openbsd
 - plan9
 - solaris
 - windows

Supported Architectures
 - 386
 - amd64
 - amd64p32
 - arm
 - arm64
 - ppc64
 - ppc64le
 - mips
 - mipsle
 - mips64
 - mips64le
 - mips64p32
 - mips64p32le
 - ppc
 - s390
 - s390x
 - sparc
 - sparc64

### Contribution

Putting everything together you can build and run your own instance by running:

    mkdir /tmp/go && cd /tmp/go
    export GOPATH=$(pwd)
    go get gopkg.in/ganggo/ganggo.v0
    cd src/gopkg.in/ganggo/ganggo.v0
    make install-deps
    cp conf/app.conf.example conf/app.conf
    
    make precompile && \
      revel run gopkg.in/ganggo/ganggo.v0

If you want to push your changes to the offical ganggo repository:

* fork the project
* create a new branch with your code changes
* create a [pull request](/ganggo/ganggo/compare)
