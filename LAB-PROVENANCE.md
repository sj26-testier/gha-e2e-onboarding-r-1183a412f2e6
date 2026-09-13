# Windows compatibility lab r-1183a412f2e6

Source: https://github.com/tartley/colorama/tree/841634ed2a0da5d5ac2d867db533da8131266cb2
License: BSD-3-Clause, retained verbatim in LICENSE.txt.
All upstream tracked files are retained. The two original workflows are preserved
byte-for-byte under .lab/upstream-workflows so they cannot trigger unintended CI.
The active test.yml changes only the upstream matrix to Python3.13/windows-latest.
No release scripts are executed. This is a bounded upstream slice, not full parity.
