#!/bin/bash

set -e

if ! command -v emcc &> /dev/null
then
  echo "Emscripten is not installed, this script only supports version 3.1.38"
  exit 1
fi
if [[ $(emcc --version) != *"3.1.38"* ]]; then
  echo "Emscripten is installed, but this must be compiled using version 3.1.38"
  echo -e "Installed Version:\n  $(emcc --version | head -n 1)"
  exit 1
fi

if ! command -v wasm2wat &> /dev/null
then
  echo "wasm2wat is not installed"
  exit 1
fi
if ! command -v wat2wasm &> /dev/null
then
  echo "wat2wasm is not installed"
  exit 1
fi


EXPORTED_FUNCTIONS="_malloc,"
EXPORTED_FUNCTIONS+="_free,"
EXPORTED_FUNCTIONS+="_uregex_open_68,"
EXPORTED_FUNCTIONS+="_uregex_close_68,"
EXPORTED_FUNCTIONS+="_uregex_start_68,"
EXPORTED_FUNCTIONS+="_uregex_end_68,"
EXPORTED_FUNCTIONS+="_uregex_find_68,"
EXPORTED_FUNCTIONS+="_uregex_findNext_68,"
EXPORTED_FUNCTIONS+="_uregex_getText_68,"
EXPORTED_FUNCTIONS+="_uregex_setText_68,"
EXPORTED_FUNCTIONS+="_uregex_replaceFirst_68,"
EXPORTED_FUNCTIONS+="_uregex_replaceAll_68,"
EXPORTED_FUNCTIONS+="_uregex_appendReplacement_68,"
EXPORTED_FUNCTIONS+="_uregex_appendTail_68,"
EXPORTED_FUNCTIONS+="_u_strFromUTF8_68,"
EXPORTED_FUNCTIONS+="_u_strToUTF8_68"

emcc \
  -s EXPORTED_FUNCTIONS="$EXPORTED_FUNCTIONS" \
  -s USE_ICU=1 \
  -s WASM=1 \
  -s TOTAL_MEMORY=64MB \
  --no-entry -o wasm/icu.wasm \
  src/file.cpp \
  -Wl,--whole-archive -licu_i18n -Wl,--no-whole-archive

wasm2wat -o wasm/icu.wat wasm/icu.wasm
sed -i 's/(export "memory" (memory 0))/(export "memory" (memory 0))(export "globalStackVar" (global 0))/' wasm/icu.wat
wat2wasm -o wasm/icu.wasm wasm/icu.wat
rm wasm/icu.wat