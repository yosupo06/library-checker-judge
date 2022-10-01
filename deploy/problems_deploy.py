#!/usr/bin/env python3

import argparse
import tempfile
import zipfile
from logging import Logger, basicConfig, getLogger
from os import getenv
from pathlib import Path

import colorlog
import grpc
import toml
from minio import Minio

from problem import Problem, find_problem_dir
import library_checker_pb2 as libpb
import library_checker_pb2_grpc

from typing import List, Optional

logger: Logger = getLogger(__name__)


def deploy_categories(stub: library_checker_pb2_grpc.LibraryCheckerServiceStub, cred_token: grpc.CallCredentials, rootdir: Path):
    categories = toml.load(open(rootdir / 'categories.toml'))
    stub.ChangeProblemCategories(libpb.ChangeProblemCategoriesRequest(
        categories=map(lambda x: libpb.ProblemCategory(
            title=x['name'],
            problems=x['problems']
        ), categories['categories'])
    ), credentials=cred_token)


def find_info_tomls(rootdir: Path, problems: List[str]) -> List[Path]:
    if len(problems) == 0:
        # enumerate all problems
        return list(filter(lambda p: not p.match(
            'test/**/info.toml'), rootdir.glob('**/info.toml')))

    tomls: List[Path] = []
    for problem_name in problems:
        problem_dir = find_problem_dir(rootdir, problem_name)
        if problem_dir is None:
            raise ValueError('Cannot find problem: {}'.format(problem_name))
        tomls.append(problem_dir / 'info.toml')
    return tomls


def connect_api_server(host: str, use_ssl: bool) -> library_checker_pb2_grpc.LibraryCheckerServiceStub:
    logger.info('connect to API server, host={} ssl={}'.format(host, use_ssl))
    if use_ssl:
        channel = grpc.secure_channel(
            host, grpc.ssl_channel_credentials())
        return library_checker_pb2_grpc.LibraryCheckerServiceStub(channel)
    else:
        channel = grpc.secure_channel(
            host, grpc.local_channel_credentials())
        return library_checker_pb2_grpc.LibraryCheckerServiceStub(channel)


def get_upload_token(stub: library_checker_pb2_grpc.LibraryCheckerServiceStub, password: str) -> grpc.CallCredentials:
    response = stub.Login(libpb.LoginRequest(name='upload', password=password))
    return grpc.access_token_call_credentials(response.token)


def get_server_problem_version(stub: library_checker_pb2_grpc.LibraryCheckerServiceStub, cred_token: grpc.CallCredentials, problem_name: str) -> Optional[str]:
    try:
        return stub.ProblemInfo(libpb.ProblemInfoRequest(
            name=problem_name), credentials=cred_token).case_version
    except grpc.RpcError as err:
        if err.code() == grpc.StatusCode.UNKNOWN:
            return None
        else:
            raise RuntimeError('unknown gRPC error:', err)


def sort_by_deploy_order(info_tomls: List[Path], stub: library_checker_pb2_grpc.LibraryCheckerServiceStub, cred_token: grpc.CallCredentials) -> List[Path]:
    tomls_new: List[Path] = []
    tomls_old: List[Path] = []

    for toml_path in info_tomls:
        probdir = toml_path.parent
        name = probdir.name
        problem = Problem(rootdir, probdir)

        new_version = problem.testcase_version()
        old_version = get_server_problem_version(stub, cred_token, name)

        if old_version is None or new_version != old_version:
            tomls_new.append(toml_path)
        else:
            tomls_old.append(toml_path)

    logger.info('prioritized problems: {}'.format(tomls_new))

    return tomls_new + tomls_old


def deploy_probelms(stub: library_checker_pb2_grpc.LibraryCheckerServiceStub, cred_token: grpc.CallCredentials, rootdir: Path, info_tomls: List[Path]):
    DEFAULT_PART_SIZE = 1 * 1000 * 1000 * 1000

    info_tomls = sort_by_deploy_order(info_tomls, stub, cred_token)

    for toml_path in info_tomls:
        probdir = toml_path.parent
        name = probdir.name
        problem = Problem(rootdir, probdir)

        new_version = problem.testcase_version()
        old_version = get_server_problem_version(stub, cred_token, name)

        logger.info('deploy : {} (version {} -> {})'.format(name,
                                                              old_version, new_version))

        problem.generate(problem.Mode.DEFAULT, None)

        title = problem.config['title']
        source_url = "https://github.com/yosupo06/library-checker-problems/tree/master/{}/{}".format(
            probdir.parent.name,
            probdir.name
        )
        timelimit = problem.config['timelimit']

        # deploy version 0
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

                # TODO: stop to remove old file manually, use auto delete
                if old_version is not None:
                    minio_client.remove_object(
                        bucket_name, old_version + '.zip')

        # deploy version 1
        # deploy test cases
        for f in sorted(probdir.glob('in/*.in')):
            path = 'v1/{}/{}/testcases/in/{}'.format(name, new_version, f.name)
            minio_client.fput_object(bucket_name, path, f, part_size=DEFAULT_PART_SIZE)
        for f in sorted(probdir.glob('out/*.out')):
            path = 'v1/{}/{}/testcases/out/{}'.format(
                name, new_version, f.name)
            minio_client.fput_object(bucket_name, path, f, part_size=DEFAULT_PART_SIZE)
        # deploy checker
        minio_client.fput_object(bucket_name, 'v1/{}/{}/checker.cpp'.format(name, new_version), probdir / 'checker.cpp', part_size=DEFAULT_PART_SIZE)
        # deploy params.h
        minio_client.fput_object(
            bucket_name, 'v1/{}/{}/include/params.h'.format(name, new_version), probdir / 'params.h', part_size=DEFAULT_PART_SIZE)
        # deploy common headers
        for f in sorted(rootdir.glob('common/*')):
            path = 'v1/{}/{}/include/{}'.format(name, new_version, f.name)
            minio_client.fput_object(bucket_name, path, f, part_size=DEFAULT_PART_SIZE)

        html = problem.gen_html()
        statement = html.statement
        stub.ChangeProblemInfo(libpb.ChangeProblemInfoRequest(
            name=name, title=title, statement=statement, time_limit=timelimit, case_version=new_version, source_url=source_url
        ), credentials=cred_token)


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

    tomls = find_info_tomls(rootdir, args.problem)

    stub = connect_api_server(args.host, args.prod)
    cred_token = get_upload_token(stub, args.api_pass)

    logger.info('connect to ObjectStorage')
    minio_client = Minio(args.minio_host,
                         access_key=args.minio_access,
                         secret_key=args.minio_secret,
                         secure=args.prod)
    bucket_name = args.minio_bucket

    if not minio_client.bucket_exists(bucket_name):
        logger.info('No bucket {}'.format(bucket_name))
        minio_client.make_bucket(bucket_name)

    deploy_categories(stub, cred_token, rootdir)
    deploy_probelms(stub, cred_token, rootdir, tomls)
