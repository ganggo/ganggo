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

All this can be installed via:

    make install-deps

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
