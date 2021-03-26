#!/usr/bin/env python3

import argparse
import hashlib
import tempfile
import zipfile
from logging import Logger, basicConfig, getLogger
from os import environ, getenv, path
from pathlib import Path

import colorlog
import grpc
import toml
from minio import Minio

from problem import Problem, find_problem_dir
import library_checker_pb2 as libpb
import library_checker_pb2_grpc

from typing import List

logger: Logger = getLogger(__name__)

if __name__ == "__main__":
    handler = colorlog.StreamHandler()
    formatter = colorlog.ColoredFormatter(
        "%(log_color)s%(asctime)s [%(levelname)s] %(message)s",
        datefmt="%H:%M:%S",
        log_colors={
            'DEBUG':    'cyan',
            'INFO':     'white',
            'WARNING':  'yellow',
            'ERROR':    'red',
            'CRITICAL': 'red,bg_white',
        })
    handler.setFormatter(formatter)
    basicConfig(
        level=getenv('LOG_LEVEL', 'INFO'),
        handlers=[handler]
    )

    parser = argparse.ArgumentParser(description='Testcase Deploy')
    parser.add_argument('root', type=Path,
                        help='the directory of librar-checker-problem')
    parser.add_argument('-p', '--problem', nargs='*',
                        help='Generate problem', default=[])
    parser.add_argument('--host', default='localhost:50051', help='Host URL')
    parser.add_argument('--prod', action='store_true',
                        help='Production Mode(use SSL)')
    parser.add_argument('--api-pass', default='password',
                        help='The password of API')
    parser.add_argument('--minio-host', default='localhost:9000',
                        help='The host of minio')
    parser.add_argument('--minio-access', default='minio',
                        help='Access key of minio')
    parser.add_argument('--minio-secret', default='miniopass',
                        help='Secret key of minio')
    parser.add_argument('--minio-bucket', default='testcase',
                        help='Bucket name of minio')

    args = parser.parse_args()
    rootdir = args.root
    tomls: List[Path] = []
    for problem_name in args.problem:
        problem_dir = find_problem_dir(rootdir, problem_name)
        if problem_dir is None:
            logger.error('Cannot find problem: {}'.format(problem_name))
            raise ValueError('Cannot find problem: {}'.format(problem_name))
        tomls.append(problem_dir / 'info.toml')
    if len(tomls) == 0:
        tomls = list(filter(lambda p: not p.match(
            'test/**/info.toml'), Path('.').glob('**/info.toml')))

    logger.info('connect to API {} ssl={}'.format(args.host, args.prod))
    if args.prod:
        channel = grpc.secure_channel(
            args.host, grpc.ssl_channel_credentials())
        stub = library_checker_pb2_grpc.LibraryCheckerServiceStub(channel)
    else:
        channel = grpc.secure_channel(
            args.host, grpc.local_channel_credentials())
        stub = library_checker_pb2_grpc.LibraryCheckerServiceStub(channel)

    api_password = args.api_pass
    response = stub.Login(libpb.LoginRequest(
        name='upload', password=api_password))
    cred_token = grpc.access_token_call_credentials(response.token)

    logger.info('connect to ObjectStorage')
    minio_client = Minio(args.minio_host,
                         access_key=args.minio_access,
                         secret_key=args.minio_secret,
                         secure=args.prod)
    bucket_name = args.minio_bucket

    if not minio_client.bucket_exists(bucket_name):
        logger.error('No bucket {}'.format(bucket_name))
        raise ValueError('No bucket {}'.format(bucket_name))

    tomls_new: List[Path] = []
    tomls_old: List[Path] = []

    for toml_path in tomls:
        probdir = toml_path.parent
        name = probdir.name
        problem = Problem(rootdir, probdir)

        new_version = problem.testcase_version()
        first_time = "FirstTime"

        try:
            old_version = stub.ProblemInfo(libpb.ProblemInfoRequest(
                name=name), credentials=cred_token).case_version
        except grpc.RpcError as err:
            if err.code() == grpc.StatusCode.UNKNOWN:
                old_version = first_time
            else:
                raise RuntimeError('Unknown gRPC error')

        if new_version != old_version:
            tomls_new.append(toml_path)
        else:
            tomls_old.append(toml_path)

    logger.info('First deploy: {}'.format(tomls_new))
    logger.info('Second deploy: {}'.format(tomls_old))

    tomls = tomls_new + tomls_old

    for toml_path in tomls:
        probdir = toml_path.parent
        name = probdir.name
        problem = Problem(rootdir, probdir)

        new_version = problem.testcase_version()
        first_time = "FirstTime"

        try:
            old_version = stub.ProblemInfo(libpb.ProblemInfoRequest(
                name=name), credentials=cred_token).case_version
        except grpc.RpcError as err:
            if err.code() == grpc.StatusCode.UNKNOWN:
                old_version = first_time
            else:
                raise RuntimeError('Unknown gRPC error')

        logger.info('Generate : {} ({} -> {})'.format(name,
                                                      old_version, new_version))

        problem.generate(problem.Mode.DEFAULT, None)

        title = problem.config['title']
        source_url = "https://github.com/yosupo06/library-checker-problems/tree/master/{}/{}".format(
            probdir.parent.name,
            probdir.name
        )
        timelimit = problem.config['timelimit']

        if new_version != old_version:
            with tempfile.NamedTemporaryFile(suffix='.zip', delete=False) as tmp:
                with zipfile.ZipFile(tmp.name, 'w', zipfile.ZIP_DEFLATED) as newzip:
                    def zip_write(filename, arcname):
                        newzip.write(filename, arcname)
                    zip_write(probdir / 'checker.cpp', arcname='checker.cpp')
                    for f in sorted(probdir.glob('in/*.in')):
                        zip_write(f, arcname=f.relative_to(probdir))
                    for f in sorted(probdir.glob('out/*.out')):
                        zip_write(f, arcname=f.relative_to(probdir))

                minio_client.fput_object(
                    bucket_name, new_version + '.zip', tmp.name, part_size=5 * 1000 * 1000 * 1000)
                if old_version != first_time:
                    minio_client.remove_object(
                        bucket_name, old_version + '.zip')

        html = problem.gen_html()
        statement = html.statement
        stub.ChangeProblemInfo(libpb.ChangeProblemInfoRequest(
            name=name, title=title, statement=statement, time_limit=timelimit, case_version=new_version, source_url=source_url
        ), credentials=cred_token)
