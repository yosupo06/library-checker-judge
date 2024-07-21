import subprocess
from pathlib import Path
from os import environ
import resource
import time

if __name__ == "__main__":
    resource.setrlimit(resource.RLIMIT_STACK, (-1, -1))

    tomls = sorted(list(filter(lambda p: not p.match('test/**/info.toml'), Path('../library-checker-problems/').glob('**/info.toml'))))

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
