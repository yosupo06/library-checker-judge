# Tools

This directory contains maintenance and operator utilities that are not part of
the deployed runtime.

Use `tools/` for small executables or scripts that are run manually by
operators, or by scheduled/one-shot automation, and whose failure does not leave
the main application in an incomplete deploy state.

Examples:

- `check-dockerfiles.sh`: Docker BuildKit build checks for Dockerfiles.
- `rejudge/`: operator CLI for queueing existing submissions for rejudge.
- `prune_gce_images.py`: housekeeping script for removing old judge VM images.

Do not put deploy/runtime components here. Components such as `migrator/`,
`uploader/`, `restapi/`, `judge/`, and `cloudrun/taskqueue-metrics/` are part of
the application deploy or runtime lifecycle, so they should remain in their
domain-specific top-level directories.
