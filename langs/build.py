#!/usr/bin/env python3
import argparse
import json
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



def normalize_keys(keys):
    # exact keys only; no aliases
    result = []
    seen = set()
    for k in keys:
        if k not in IMAGES:
            raise SystemExit(f"Unknown image key: {k}")
        if k not in seen:
            result.append(k)
            seen.add(k)
    return result


def build_one(key, *, tag_prefix=""):
    suffix, tag = IMAGES[key]
    dockerfile = SCRIPT_DIR / f"Dockerfile.{suffix}"
    full_tag = f"{tag_prefix}{tag}" if tag_prefix else tag
    cmd = [
        "docker", "build",
        "-t", full_tag,
        "-f", str(dockerfile),
        str(SCRIPT_DIR),
    ]
    print("+", " ".join(cmd), flush=True)
    subprocess.run(cmd, check=True)
    inspect = subprocess.run(
        ["docker", "image", "inspect", full_tag, "--format", "{{.Size}}"],
        check=True,
        capture_output=True,
        text=True,
    )
    size = int(inspect.stdout.strip())
    return {
        "key": key,
        "dockerfile": f"Dockerfile.{suffix}",
        "tag": full_tag,
        "size_bytes": size,
    }


def main(argv):
    parser = argparse.ArgumentParser(description="Build language Docker images")
    parser.add_argument(
        "images",
        nargs="*",
        help=(
            "Image keys to build (default: all). "
            "Examples: gcc python3 | all."
        ),
    )
    parser.add_argument(
        "--tag-prefix",
        default="",
        help="Prefix to prepend to image tags during build",
    )
    parser.add_argument(
        "--output-json",
        type=Path,
        help="Path to write build metadata (JSON).",
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
        return 0

    if not args.images or (len(args.images) == 1 and args.images[0] == "all"):
        keys = list(IMAGES.keys())
    else:
        keys = normalize_keys(args.images)

    metadata = []
    for key in keys:
        metadata.append(
            build_one(
                key,
                tag_prefix=args.tag_prefix,
            )
        )
    if args.output_json:
        args.output_json.parent.mkdir(parents=True, exist_ok=True)
        args.output_json.write_text(json.dumps(metadata, indent=2) + "\n")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
