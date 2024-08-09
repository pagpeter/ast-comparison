# JavaScript AST library comparison

This repository aims to compare the performance of multiple different JavaScript AST libraries written in different languages. The tests are not the usual TypeScript to JavaScript things, but custom transformations for deobfuscating obfuscated JavaScript.

It is not finished.

The first three libraries I aim to test are:

- [go-fAST](https://github.com/T14Raptor/go-fAST)
- [swc](https://crates.io/crates/swc)
- [babel](https://www.npmjs.com/package/@babel/core)

In the process, it also shows examples on how the libraries can be used.

## Credits

Most of the stuff was written by me. However, i'd like to thank

- @T14Raptor for the go-fAST example
- @rsa2048 for some speedups in the swc test
