#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: prune_gce_images.sh --project <PROJECT_ID> --family <IMAGE_FAMILY> [options]

Options:
  --keep <N>           Number of latest images to keep (default: 10)
  --min-age-days <N>   Minimum age in days before an image can be deleted (default: 14)
  --dry-run            Only print the images that would be deleted
  -h, --help           Show this help message
USAGE
}

PROJECT=""
FAMILY=""
KEEP=10
MIN_AGE_DAYS=14
DRY_RUN=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --project)
      PROJECT="$2"
      shift 2
      ;;
    --family)
      FAMILY="$2"
      shift 2
      ;;
    --keep)
      KEEP="$2"
      shift 2
      ;;
    --min-age-days)
      MIN_AGE_DAYS="$2"
      shift 2
      ;;
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [[ -z "$PROJECT" || -z "$FAMILY" ]]; then
  echo "--project and --family are required" >&2
  usage >&2
  exit 1
fi

if ! [[ "$KEEP" =~ ^[0-9]+$ && "$MIN_AGE_DAYS" =~ ^[0-9]+$ ]]; then
  echo "--keep and --min-age-days must be non-negative integers" >&2
  exit 1
fi

now_epoch=$(date +%s)
min_age_seconds=$(( MIN_AGE_DAYS * 24 * 60 * 60 ))

mapfile -t images < <(gcloud compute images list \
  --project="$PROJECT" \
  --filter="family=$FAMILY" \
  --no-standard-images \
  --format="value(name,creationTimestamp)" \
  --sort-by="~creationTimestamp")

total=${#images[@]}
if (( total == 0 )); then
  echo "No images found for family '$FAMILY' in project '$PROJECT'."
  exit 0
fi

echo "Found $total images for family '$FAMILY' in project '$PROJECT'. Keeping the latest $KEEP images."

if (( total <= KEEP )); then
  echo "Nothing to prune."
  exit 0
fi

pruned_any=false
for (( i=KEEP; i<total; i++ )); do
  entry=${images[$i]}

  IFS=$'\t' read -r name timestamp <<<"$entry"
  if [[ -z "$name" || -z "$timestamp" ]]; then
    continue
  fi

  if ! image_epoch=$(date --date="$timestamp" +%s 2>/dev/null); then
    echo "Skipping $name (failed to parse timestamp '$timestamp')." >&2
    continue
  fi
  age_seconds=$(( now_epoch - image_epoch ))

  if (( age_seconds < min_age_seconds )); then
    echo "Skipping $name (only $(( age_seconds / 86400 )) days old)."
    continue
  fi

  pruned_any=true
  if [[ "$DRY_RUN" == true ]]; then
    echo "[DRY RUN] Would delete image $name"
  else
    echo "Deleting image $name"
    gcloud compute images delete "$name" --project="$PROJECT" --quiet
  fi
done

if [[ "$DRY_RUN" == true ]]; then
  echo "Dry run completed."
elif [[ "$pruned_any" == false ]]; then
  echo "No images matched pruning criteria."
fi
