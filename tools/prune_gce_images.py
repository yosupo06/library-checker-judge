#!/usr/bin/env python3
"""Prune old GCE images while keeping the newest ones."""

from __future__ import annotations

import argparse
import subprocess
import sys
from dataclasses import dataclass
from datetime import datetime, timezone, timedelta
from typing import Iterable, List


@dataclass
class ImageEntry:
    name: str
    timestamp: str

    def parsed_time(self) -> datetime | None:
        value = self.timestamp.strip()
        if not value:
            return None
        # gcloud emits RFC3339 timestamps, ensure timezone aware
        sanitized = value.replace("Z", "+00:00")
        try:
            dt = datetime.fromisoformat(sanitized)
        except ValueError:
            return None
        if dt.tzinfo is None:
            dt = dt.replace(tzinfo=timezone.utc)
        return dt.astimezone(timezone.utc)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Delete old images from a GCE image family",
    )
    parser.add_argument("--project", required=True, help="GCP project ID")
    parser.add_argument("--family", required=True, help="GCE image family name")
    parser.add_argument("--keep", type=int, default=10,
                        help="Number of most recent images to keep (default: 10)")
    parser.add_argument("--min-age-days", type=int, default=14,
                        help="Minimum age in days before deleting an image (default: 14)")
    parser.add_argument("--dry-run", action="store_true",
                        help="Only print images that would be deleted")
    return parser.parse_args()


def validate_args(args: argparse.Namespace) -> None:
    if args.keep < 0:
        sys.exit("--keep must be a non-negative integer")
    if args.min_age_days < 0:
        sys.exit("--min-age-days must be a non-negative integer")


def run_gcloud_list(project: str, family: str) -> List[ImageEntry]:
    cmd = [
        "gcloud",
        "compute",
        "images",
        "list",
        f"--project={project}",
        f"--filter=family={family}",
        "--no-standard-images",
        "--format=value(name,creationTimestamp)",
        "--sort-by=~creationTimestamp",
    ]
    try:
        result = subprocess.run(
            cmd,
            check=True,
            text=True,
            capture_output=True,
        )
    except subprocess.CalledProcessError as exc:
        sys.stderr.write(exc.stderr)
        raise SystemExit(exc.returncode)

    entries: List[ImageEntry] = []
    for raw_line in result.stdout.splitlines():
        line = raw_line.strip()
        if not line:
            continue
        if "\t" in line:
            name, timestamp = line.split("\t", 1)
        else:
            parts = line.split(None, 1)
            if len(parts) == 1:
                name, timestamp = parts[0], ""
            else:
                name, timestamp = parts[0], parts[1]
        entries.append(ImageEntry(name=name.strip(), timestamp=timestamp.strip()))
    return entries


def prune_images(entries: Iterable[ImageEntry], keep: int,
                 min_age_days: int, dry_run: bool,
                 project: str, family: str) -> None:
    entries_list = list(entries)
    total = len(entries_list)
    if total == 0:
        print(f"No images found for family '{family}' in project '{project}'.")
        return

    print(
        f"Found {total} images for family '{family}' in project '{project}'. "
        f"Keeping the latest {keep} images."
    )
    if total <= keep:
        print("Nothing to prune.")
        return

    now = datetime.now(timezone.utc)
    cutoff = now - timedelta(days=min_age_days)
    pruned_any = False

    for entry in entries_list[keep:]:
        created_at = entry.parsed_time()
        if created_at is None:
            print(f"Skipping {entry.name} (invalid timestamp '{entry.timestamp}').", file=sys.stderr)
            continue
        if created_at > cutoff:
            age_days = (now - created_at).days
            print(f"Skipping {entry.name} (only {age_days} days old).")
            continue

        pruned_any = True
        if dry_run:
            print(f"[DRY RUN] Would delete image {entry.name}")
            continue

        print(f"Deleting image {entry.name}")
        delete_cmd = [
            "gcloud",
            "compute",
            "images",
            "delete",
            entry.name,
            f"--project={project}",
            "--quiet",
        ]
        try:
            subprocess.run(delete_cmd, check=True)
        except subprocess.CalledProcessError as exc:
            print(f"Failed to delete {entry.name}: {exc}", file=sys.stderr)
            raise SystemExit(exc.returncode)

    if dry_run:
        print("Dry run completed.")
    elif not pruned_any:
        print("No images matched pruning criteria.")


def main() -> None:
    args = parse_args()
    validate_args(args)
    entries = run_gcloud_list(args.project, args.family)
    prune_images(
        entries,
        args.keep,
        args.min_age_days,
        args.dry_run,
        args.project,
        args.family,
    )


if __name__ == "__main__":
    main()
