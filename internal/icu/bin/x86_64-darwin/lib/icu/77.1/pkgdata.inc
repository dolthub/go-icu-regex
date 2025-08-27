GENCCODE_ASSEMBLY_TYPE=-a gcc-darwin
SO=dylib
SOBJ=dylib
A=a
LIBPREFIX=lib
LIB_EXT_ORDER=.77.1.dylib
COMPILE=clang-15 --target=x86_64-darwin --sysroot=/opt/cross/darwin-sysroot -mmacosx-version-min=12.0 -DU_ALL_IMPLEMENTATION -DU_ATTRIBUTE_DEPRECATED= -DUCONFIG_NO_LEGACY_CONVERSION=1   -DUCONFIG_NO_BREAK_ITERATION=1   -DUCONFIG_NO_COLLATION=1   -DUCONFIG_NO_FORMATTING=1   -DUCONFIG_NO_TRANSLITERATION=1   -DICU_DATA_DIR=""    -DU_HAVE_STRTOD_L=1 -DU_HAVE_XLOCALE_H=1  -O2  -std=c11 -Wall -pedantic -Wshadow -Wpointer-arith -Wmissing-prototypes -Wwrite-strings   -fno-common -c
LIBFLAGS=-I//include  -dynamic
GENLIB=clang-15 --target=x86_64-darwin --sysroot=/opt/cross/darwin-sysroot -mmacosx-version-min=12.0 -dynamiclib -dynamic -O2  -std=c11 -Wall -pedantic -Wshadow -Wpointer-arith -Wmissing-prototypes -Wwrite-strings    
LDICUDTFLAGS=
LD_SONAME=-Wl,-compatibility_version -Wl,77 -Wl,-current_version -Wl,77.1 -install_name  
RPATH_FLAGS=
BIR_LDFLAGS=
AR=x86_64-darwin-ar
ARFLAGS=r -c
RANLIB=x86_64-darwin-ranlib
INSTALL_CMD=/usr/bin/install -c
