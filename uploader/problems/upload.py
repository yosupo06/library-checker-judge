import subprocess
from pathlib import Path
from os import environ
import resource
import time

if __name__ == "__main__":
    resource.setrlimit(resource.RLIMIT_STACK, (-1, -1))

    tomls = sorted(list(filter(lambda p: not p.match('test/**/info.toml'), Path('../library-checker-problems/').glob('**/info.toml'))))

    # optional sharding for GitHub Actions matrix (split across runners)
    ST = int(environ.get("SHARD_TOTAL", "1"))
    SI = int(environ.get("SHARD_INDEX", "0"))
    if ST < 1:
        raise ValueError(f"SHARD_TOTAL must be >= 1, got {ST}")
    if SI < 0 or SI >= ST:
        raise ValueError(f"SHARD_INDEX must be in [0,{ST-1}], got {SI}")
    tomls = [t for i, t in enumerate(tomls) if i % ST == SI]

    DISCORD_WEBHOOK = environ["DISCORD_WEBHOOK"]
    FORCE_UPLOAD = environ["FORCE_UPLOAD"]

    subprocess.run(
        ["./uploader"] +
        ["-discordwebhook", DISCORD_WEBHOOK] +
        ["-dir", "../library-checker-problems"] +
        (["-force"] if FORCE_UPLOAD == "true" else []) +
        [str(toml.absolute()) for toml in tomls],
        check=True
    )
