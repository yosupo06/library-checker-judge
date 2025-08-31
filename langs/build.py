#!/usr/bin/env python3
import argparse
import subprocess
import sys
from pathlib import Path


SCRIPT_DIR = Path(__file__).resolve().parent

# key -> (dockerfile suffix, image tag)
IMAGES = {
    "gcc": ("GCC", "library-checker-images-gcc"),
    "ldc": ("LDC", "library-checker-images-ldc"),  # D (ldc)
    "python3": ("PYTHON3", "library-checker-images-python3"),
    "haskell": ("HASKELL", "library-checker-images-haskell"),
    "csharp": ("CSHARP", "library-checker-images-csharp"),
    "rust": ("RUST", "library-checker-images-rust"),
    "java": ("JAVA", "library-checker-images-java"),
    "pypy": ("PYPY", "library-checker-images-pypy"),
    "golang": ("GOLANG", "library-checker-images-golang"),
    "lisp": ("LISP", "library-checker-images-lisp"),
    "crystal": ("CRYSTAL", "library-checker-images-crystal"),
    "ruby": ("RUBY", "library-checker-images-ruby"),
    "swift": ("SWIFT", "library-checker-images-swift"),
}

# Aliases for convenience
ALIASES = {
    "d": "ldc",
    "go": "golang",
    "pypy3": "pypy",
    # cpp-related languages use gcc image but are not images themselves
    # expose a helper so users can say `cpp` and we still build gcc
    "cpp": "gcc",
}


def normalize_keys(keys):
    result = []
    for k in keys:
        norm = ALIASES.get(k, k)
        if norm not in IMAGES:
            raise SystemExit(f"Unknown image key: {k}")
        result.append(norm)
    # de-duplicate while preserving order
    seen = set()
    deduped = []
    for k in result:
        if k not in seen:
            deduped.append(k)
            seen.add(k)
    return deduped


def build_one(key):
    suffix, tag = IMAGES[key]
    dockerfile = SCRIPT_DIR / f"Dockerfile.{suffix}"
    cmd = [
        "docker", "build",
        "-t", tag,
        "-f", str(dockerfile),
        str(SCRIPT_DIR),
    ]
    print("+", " ".join(cmd), flush=True)
    subprocess.run(cmd, check=True)


def main(argv):
    parser = argparse.ArgumentParser(description="Build language Docker images")
    parser.add_argument(
        "images",
        nargs="*",
        help=(
            "Image keys to build (default: all). "
            "Examples: gcc python3 | all. Aliases: d->ldc, go->golang, pypy3->pypy, cpp->gcc"
        ),
    )
    parser.add_argument(
        "--list",
        action="store_true",
        help="List available image keys and exit",
    )
    args = parser.parse_args(argv)

    if args.list:
        print("Available images:")
        for k in IMAGES:
            print(" -", k)
        print("Aliases:")
        for a, t in ALIASES.items():
            print(f" - {a} -> {t}")
        return 0

    if not args.images or (len(args.images) == 1 and args.images[0] == "all"):
        keys = list(IMAGES.keys())
    else:
        keys = normalize_keys(args.images)

    for key in keys:
        build_one(key)
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
