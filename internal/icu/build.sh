#!/bin/bash

set -eo pipefail

docker run --rm -v `pwd`:/out ubuntu /bin/bash -c '
set -eo pipefail

apt-get update
DEBIAN_FRONTEND=noninteractive apt-get install -y p7zip-full pigz curl xz-utils mingw-w64 clang-15 build-essential

# For the linux-musl and darwin builds, we will use our own crosstools.
cd /
curl -o optcross.tar.xz https://dolthub-tools.s3.us-west-2.amazonaws.com/optcross/"$(uname -m)"-linux_20250327_0.0.3.tar.xz
tar Jxf optcross.tar.xz
export PATH=/opt/cross/bin:"$PATH"
rm optcross.tar.xz

# Start by making our working directories and fetching the ICU source.
mkdir /source /build /install

cd /source
curl -OL https://github.com/unicode-org/icu/releases/download/release-77-1/icu4c-77_1-src.tgz
tar zxvf icu4c-77_1-src.tgz

# First build a native copy of ICU, to use for the cross builds.
cd /build
mkdir native
cd native
/source/icu/source/runConfigureICU Linux
make -j8

# We have 5 targets to build for:
# * x86_64-w64-mingw32
# * x86_64-linux-musl
# * aarch64-linux-musl
# * x86_64-darwin
# * aarch64-darwin

CONFIG_CPPFLAGS="\
  -DUCONFIG_NO_LEGACY_CONVERSION=1 \
  -DUCONFIG_NO_BREAK_ITERATION=1 \
  -DUCONFIG_NO_COLLATION=1 \
  -DUCONFIG_NO_FORMATTING=1 \
  -DUCONFIG_NO_TRANSLITERATION=1 \
  -DICU_DATA_DIR="'\''\"\"'\''" \
"

# The first three cross builds look similar to each other:
for TARGET in x86_64-w64-mingw32 x86_64-linux-musl aarch64-linux-musl; do
    cd /build
    mkdir $TARGET
    cd $TARGET

    AR=$TARGET-ar \
    CC=$TARGET-gcc \
    CXX=$TARGET-g++ \
    RANLIB=$TARGET-ranlib \
    CPPFLAGS="$CONFIG_CPPFLAGS" \
        /source/icu/source/runConfigureICU MinGW \
            --host=$TARGET \
            --with-cross-build=/build/native \
            --enable-static \
            --disable-shared \
            --disable-tools \
            --disable-tests \
            --disable-extras \
            --with-data-packaging=archive \
            --prefix=/
    make -j8
    make install DESTDIR=/install/$TARGET
done

# The darwin cross builds look a little different.

for TARGET in aarch64 x86_64; do
    cd /build
    mkdir $TARGET-darwin
    cd $TARGET-darwin

    AR=$TARGET-darwin-ar \
    CC="clang-15 --target=$TARGET-darwin --sysroot=/opt/cross/darwin-sysroot -mmacosx-version-min=12.0" \
    CXX="clang++-15 --target=$TARGET-darwin --sysroot=/opt/cross/darwin-sysroot -mmacosx-version-min=12.0 -stdlib=libc++" \
    RANLIB=$TARGET-darwin-ranlib \
    CPPFLAGS="$CONFIG_CPPFLAGS" \
        /source/icu/source/runConfigureICU macOS \
            --host=$TARGET-apple-darwin \
            --with-cross-build=/build/native \
            --enable-static \
            --disable-shared \
            --disable-tools \
            --disable-tests \
            --disable-extras \
            --with-data-packaging=archive \
            --prefix=/
    make -j8
    make install DESTDIR=/install/$TARGET-darwin
done

# Create a tar file with the exported files.

cd /install
tar cf exported.tar -C aarch64-linux-musl include
tar rf exported.tar */lib
mv exported.tar /out
'
