FROM golang:1.9-stretch

RUN curl -sL https://deb.nodesource.com/setup_10.x | bash -
RUN apt-get install -y nodejs

# armhf c cross compiler
RUN apt-get install -y gcc-arm-linux-gnueabihf

# windows 64 and 32 c/c++ cross compiler
RUN apt-get install -y gcc-mingw-w64-x86-64 gcc-mingw-w64-i686
RUN apt-get install -y g++-mingw-w64-x86-64 g++-mingw-w64-i686
