import subprocess
from pathlib import Path
from os import environ
import resource

if __name__ == "__main__":
    resource.setrlimit(resource.RLIMIT_STACK, (-1, -1))

    index = int(environ["CLOUD_RUN_TASK_INDEX"])
    count = int(environ["CLOUD_RUN_TASK_COUNT"])

    tomls = sorted(list(filter(lambda p: not p.match('test/**/info.toml'), Path('../library-checker-problems/').glob('**/info.toml'))))[index::count]

    print("tomls: ", tomls)

    PG_USER = environ["PG_USER"]
    PG_PASS = environ["PG_PASS"]
    PG_TABLE = environ["PG_TABLE"]
    MINIO_HOST = environ["MINIO_HOST"]
    MINIO_ID = environ["MINIO_ID"]
    MINIO_SECRET = environ["MINIO_SECRET"]
    MINIO_BUCKET = environ["MINIO_BUCKET"]

    for toml in tomls:
        subprocess.run(
            ["../library-checker-problems/generate.py",
                "--only-html", str(toml.absolute())],
            check=True
        )

        subprocess.run(
            ["./uploader"] +
            ["-pguser", PG_USER] +
            ["-pgpass", PG_PASS] +
            ["-pgtable", PG_TABLE] +
            ["-miniohost", MINIO_HOST] +
            ["-minioid", MINIO_ID] +
            ["-miniokey", MINIO_SECRET] +
            ["-miniobucket", MINIO_BUCKET] +
            ["-dir", "../library-checker-problems"] +
            ["-tls"] +
            ["-toml", str(toml.absolute())],
            check=True
        )

        subprocess.run(
            ["../library-checker-problems/generate.py",
                "--clean", str(toml.absolute())],
            check=True
        )
