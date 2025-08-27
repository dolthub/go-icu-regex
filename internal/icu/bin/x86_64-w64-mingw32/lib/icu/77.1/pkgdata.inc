GENCCODE_ASSEMBLY_TYPE=-a gcc-mingw64
SO=dll
SOBJ=dll
A=a
LIBPREFIX=s
LIB_EXT_ORDER=77.dll
COMPILE=x86_64-w64-mingw32-gcc -DUCONFIG_NO_LEGACY_CONVERSION=1   -DUCONFIG_NO_BREAK_ITERATION=1   -DUCONFIG_NO_COLLATION=1   -DUCONFIG_NO_FORMATTING=1   -DUCONFIG_NO_TRANSLITERATION=1   -DICU_DATA_DIR=""    -DHAVE_DLOPEN=0 -DU_HAVE_MMAP=0 -DU_HAVE_STRTOD_L=0  -DU_STATIC_IMPLEMENTATION -DWINVER=0x0601 -D_WIN32_WINNT=0x0601 -DU_ALL_IMPLEMENTATION -DU_ATTRIBUTE_DEPRECATED= -O3  -std=c11 -Wall -pedantic -Wshadow -Wpointer-arith -Wmissing-prototypes -Wwrite-strings -mthreads  -c
LIBFLAGS=-I//include -DPIC 
GENLIB=x86_64-w64-mingw32-gcc -O3  -std=c11 -Wall -pedantic -Wshadow -Wpointer-arith -Wmissing-prototypes -Wwrite-strings -mthreads   -shared  -Wl,--enable-auto-import -Wl,--out-implib=./liball.dll.a
LDICUDTFLAGS=
LD_SONAME=
RPATH_FLAGS=
BIR_LDFLAGS=
AR=x86_64-w64-mingw32-ar
ARFLAGS=r
RANLIB=x86_64-w64-mingw32-ranlib
INSTALL_CMD=/usr/bin/install -c
