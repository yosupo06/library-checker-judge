# Swift judge image

The image uses Swift 6.4.2 with Collections 1.7.6 and Algorithms 1.2.1.
Numerics 1.1.1 is pinned as Algorithms' transitive dependency.

SwiftPM builds a static library when the Docker image is built. The final image
contains the archive, Swift modules, Numerics' Clang module and headers, and the
packages' license files. Submissions can use `import Collections`, individual
collection modules such as `HeapModule`, and `import Algorithms`.

Submissions are still compiled directly with `swiftc -O`; no package resolution
or network access occurs during judging. The default Swift language mode is
preserved (Swift 5); upgrading the compiler does not force submitted sources
into Swift 6 concurrency checking. No experimental Collections traits are enabled.

Build and verify from the repository root:

```sh
python3 langs/build.py swift
cd executor
go test -v -count=1 -run '^(TestAllLangsAplusb|TestSwiftLibraries)$/^swift$' .
```

The executor tests require the same Docker/crun setup as Test Executor CI.
The A+B test covers an ordinary submission, and TestSwiftLibraries covers
imports and execution within the judge's 30-second/1-GiB compile budget.
