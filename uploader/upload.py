import subprocess
from pathlib import Path
from os import environ
import resource

if __name__ == "__main__":
    resource.setrlimit(resource.RLIMIT_STACK, (-1, -1))

    tomls = sorted(list(filter(lambda p: not p.match('test/**/info.toml'), Path('../library-checker-problems/').glob('**/info.toml'))))

    PG_USER = environ["PG_USER"]
    PG_PASS = environ["PG_PASS"]
    PG_TABLE = environ["PG_TABLE"]
    MINIO_HOST = environ["MINIO_HOST"]
    MINIO_ID = environ["MINIO_ID"]
    MINIO_SECRET = environ["MINIO_SECRET"]
    MINIO_BUCKET = environ["MINIO_BUCKET"]
    DISCORD_WEBHOOK = environ["DISCORD_WEBHOOK"]
    FORCE_UPLOAD = environ["FORCE_UPLOAD"]

    print("force: ", FORCE_UPLOAD)
    subprocess.run(
        ["./uploader"] +
        ["-pguser", PG_USER] +
        ["-pgpass", PG_PASS] +
        ["-pgtable", PG_TABLE] +
        ["-miniohost", MINIO_HOST] +
        ["-minioid", MINIO_ID] +
        ["-miniokey", MINIO_SECRET] +
        ["-miniobucket", MINIO_BUCKET] +
        ["-discordwebhook", DISCORD_WEBHOOK] +
        ["-dir", "../library-checker-problems"] +
        (["-force"] if FORCE_UPLOAD == "true" else []) +
        ["-tls"] +
        [str(toml.absolute()) for toml in tomls],
        check=True
    )
