#!/usr/bin/env python3
# Copyright 2019 Kohei Morita
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


import json
import shutil
import sys
import tempfile
import traceback
import zipfile
import datetime
from copy import deepcopy
from datetime import datetime
from logging import Logger, basicConfig, getLogger
from os import environ, getenv
from pathlib import Path
from shutil import copy, rmtree
from subprocess import (DEVNULL, PIPE, STDOUT, CalledProcessError, Popen,
                        TimeoutExpired, check_call, run)
from time import sleep

import psycopg2
import toml
from termcolor import colored, cprint

from judgeinside import Judgement, Result

basicConfig(
    level=getenv('LOG_LEVEL', 'DEBUG'),
    format="%(asctime)s %(levelname)s %(name)s :%(message)s"
)
logger: Logger = getLogger(__name__)

class Problem:
    id: str
    testhash: str
    timelimit: float

    def __init__(self, conn, id: str):
        self.id = id
        with conn.cursor() as cursor:
            if cursor.execute('select testhash, timelimit from problems where name = %s', (id, )) == 0:
                raise Exception()
            prob = cursor.fetchone()
            self.testhash = prob[0]
            self.timelimit = prob[1]

    def fetchcase(self, conn, zippath: Path):
        if zippath.exists():
            return
        logger.info('Nothing {}, fetching...'.format(zippath))
        with conn.cursor() as cursor:
            if cursor.execute('select testzip from problems where name = %s', (self.id, )) == 0:
                raise Exception()
            zipdata = cursor.fetchone()[0]
            with open(zippath, 'wb') as f:
                f.write(zipdata)

class Submission:
    id: int
    pid: str
    lang: str
    source: str

    def __init__(self, conn, id: int):
        self.id = id
        with conn.cursor() as cursor:
            if cursor.execute('select problem_name, lang, source from submissions where id = %s', (id, )) == 0:
                raise Exception()
            submission = cursor.fetchone()
            self.pid = submission[0]
            self.lang = submission[1]
            self.source = submission[2]

    def ref_status(self, conn, status: str):
        with conn.cursor() as cursor:
            cursor.execute('update submissions set status = %s where id = %s',
                        (status, self.id))
            conn.commit()
        self.ref_ping(conn)

    def ref_ping(self, conn):
        with conn.cursor() as cursor:
            cursor.execute('update submissions set judge_ping = %s where id = %s',
                        (datetime.now(), self.id))
            conn.commit()

    def clear_ping(self, conn):
        with conn.cursor() as cursor:
            cursor.execute('update submissions set judge_ping = NULL where id = %s',
                        (self.id, ))
            conn.commit()


def fetchdata(conn, problem: Problem):
    zippath = workdir / 'cases-{}.zip'.format(problem.testhash)
    problem.fetchcase(conn, zippath)

    logger.info('Extract zip file')

    if (workdir / 'checker.cpp').exists():
        (workdir / 'checker.cpp').unlink()
    indir = workdir / 'in'
    outdir = workdir / 'out'
    if indir.exists():
        shutil.rmtree(indir)
    if outdir.exists():
        shutil.rmtree(outdir)

    with zipfile.ZipFile(zippath, 'r') as f:
        f.extractall(workdir)


def judge(conn, submission: Submission):
    # WJ -> Fetching
    with conn.cursor() as cursor:
        cursor.execute("update submissions set (status, judge_ping) = (%s, %s) where id = %s and status = 'WJ'",
                          ('Fetching', datetime.now(), submission.id))
        if not cursor.rowcount or cursor.rowcount <= 0:
            return
        conn.commit()

    # Delete judge status
    with conn.cursor() as cursor:
        cursor.execute(
            'delete from submission_testcase_results where submission = %s', (submission.id,))
        conn.commit()

    logger.info('Judge start submission id = {}'.format(submission.id))

    logger.info('Fetch data from SQL')

    problem = Problem(conn, submission.pid)
    fetchdata(conn, problem)

    judgement = Judgement(submission.lang)

    # write source
    with open(workdir / judgement.lang.source, 'w') as f:
        f.write(submission.source)

    submission.ref_status(conn, 'Compiling')

    if not judgement.compile_checker():
        submission.ref_status(conn, 'ICE')
        return

    if not judgement.compile():
        submission.ref_status(conn, 'CE')
        return

    submission.ref_status(conn, 'Executing')

    all_result = Result('AC')

    def refresh(name: str, result: Result):
        submission.ref_ping(conn)
        with conn.cursor() as cursor:
            cursor.execute('''insert into submission_testcase_results
                              (submission, testcase, status, time, memory)
                              values (%s, %s, %s, %s, %s)''',
                           (submission.id, name, result.status, result.time, result.memory))
            conn.commit()
        if result.status != 'AC':
            all_result.status = result.status
        all_result.time = max(all_result.time, result.time)
        all_result.memory = max(all_result.memory, result.memory)

    judgement.judge(refresh, problem.timelimit / 1000)

    with conn.cursor() as cursor:
        cursor.execute('update submissions set status = %s, max_time = %s, max_memory = %s where id = %s',
                       (all_result.status, all_result.time, all_result.memory, submission.id))
        conn.commit()

    logger.info('End judge')
    submission.clear_ping(conn)


if __name__ == "__main__":
    logger.info('Launch judge.py')


    logger.info('Make workdir')
    workdir = Path.cwd() / 'work'

    if workdir.exists():
        rmtree(workdir)
    workdir.mkdir()
    shutil.copy('testlib.h', workdir / 'testlib.h')


    logger.info('Connect SQL')
    hostname = environ.get('POSTGRE_HOST', '127.0.0.1')
    port = int(environ.get('POSTGRE_PORT', '5432'))
    user = environ.get('POSTGRE_USER', 'postgres')
    password = environ.get('POSTGRE_PASS', 'passwd')
    conn = psycopg2.connect(
        host=hostname,
        port=port,
        user=user,
        password=password,
        database='librarychecker'
    )
    while True:
        sleep(1)
        sql = 'select id from queue'
        res = None
        with conn.cursor() as cursor:
            cursor.execute('select id, submission from tasks limit 1')
            res = cursor.fetchone()
            if res == None:
                continue
            cursor.execute('delete from tasks where id = %s', (res[0],))
            conn.commit()

        submission = Submission(conn, res[1])
        try:
            judge(conn, submission)
        except Exception as e:
            logger.exception(e)
            with conn.cursor() as cursor:
                cursor.execute('update submissions set status = %s where id = %s',
                               ('IE', submission.id))
                conn.commit()
