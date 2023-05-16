# ICU Regular Expressions in Go

The [ICU library](https://github.com/unicode-org/icu) is used in MySQL to parse regular expressions.
Go's built-in regular expressions follow a different standard than ICU, and thus can cause inconsistencies when attempting to match MySQL's behavior.
These inconsistencies would hopefully result in an error (prompting user intervention), but may silently return unexpected results, raising no alarm when data is being modified in unexpected ways.

To get around this, we've implemented the necessary ICU functions by compiling them into a [WebAssembly](https://webassembly.org/) module, and running the module using the [wazero](https://github.com/tetratelabs/wazero) library.
Although this approach does come with a performance penalty, this allows for implementing packages to retain cross-compilation support, as CGo is not invoked due to this package.

## Building

To make modifications to the compiled WASM module, we've included a [build script](icu/build.sh).
The requirements are as follows:

* Emscripten v3.1.38
* wasm2wat
* wat2wasm

Other Emscripten versions may compile just fine, however they have not been tested, and thus we restrict compilation to only the tested version.
This also means that the ICU library is version [68.1](https://github.com/unicode-org/icu/tree/5d81f6f9a0edc47892a1d2af7024f835b47deb82), as that is the only version that our supported version of Emscripten has ported.
Both `wasm2wat` and `wat2wasm` exist to expose the global stack variable, as not all platforms will expose the variable.
None of the exposed functions require [ICU's data](https://unicode-org.github.io/icu/userguide/icu_data/), thus it has been excluded to save on space and memory usage.
MySQL, although collation aware (and in spite of what the documentation may suggest), does not make use of any collation functionality in the context of regular expressions.

## Notes

Due to the high startup-cost of the WASM runtime, this package _enforces_ that all Regex objects are closed before being dereferenced.
If any Regex objects are dereferenced before being closed, then a panic will occur at some non-deterministic point in the future.