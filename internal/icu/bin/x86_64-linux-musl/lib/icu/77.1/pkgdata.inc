GENCCODE_ASSEMBLY_TYPE=-a gcc
SO=so
SOBJ=so
A=a
LIBPREFIX=lib
LIB_EXT_ORDER=.77.1
COMPILE=x86_64-linux-musl-gcc -DUCONFIG_NO_LEGACY_CONVERSION=1   -DUCONFIG_NO_BREAK_ITERATION=1   -DUCONFIG_NO_COLLATION=1   -DUCONFIG_NO_FORMATTING=1   -DUCONFIG_NO_TRANSLITERATION=1   -DICU_DATA_DIR=""  -ffunction-sections -fdata-sections -D_REENTRANT  -DU_HAVE_ELF_H=1 -DU_HAVE_STRTOD_L=1 -DU_HAVE_XLOCALE_H=0  -DU_ALL_IMPLEMENTATION -DU_ATTRIBUTE_DEPRECATED= -O3  -std=c11 -Wall -pedantic -Wshadow -Wpointer-arith -Wmissing-prototypes -Wwrite-strings   -c
LIBFLAGS=-I//include -DPIC -fPIC
GENLIB=x86_64-linux-musl-gcc -O3  -std=c11 -Wall -pedantic -Wshadow -Wpointer-arith -Wmissing-prototypes -Wwrite-strings   -Wl,--gc-sections  -shared -Wl,-Bsymbolic
LDICUDTFLAGS=-nodefaultlibs -nostdlib
LD_SONAME=-Wl,-soname -Wl,
RPATH_FLAGS=
BIR_LDFLAGS=-Wl,-Bsymbolic
AR=x86_64-linux-musl-ar
ARFLAGS=r
RANLIB=x86_64-linux-musl-ranlib
INSTALL_CMD=/usr/bin/install -c
