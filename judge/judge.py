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


from pathlib import Path
from os import getenv, environ
from shutil import copy, rmtree
import json
import shutil
import sys
import tempfile
import traceback
import zipfile
from datetime import datetime
from logging import basicConfig, getLogger, Logger
from subprocess import (DEVNULL, PIPE, CalledProcessError, Popen,
                        TimeoutExpired, check_call, run)
from time import sleep

import psycopg2
from termcolor import colored, cprint

basicConfig(
    level=getenv('LOG_LEVEL', 'DEBUG'),
    format="%(asctime)s %(levelname)s %(name)s :%(message)s"
)
logger: Logger = getLogger(__name__)

logger.info('Launch judge.py')


class UnknownTypeFile(Exception):
    def __init__(self, message):
        super().__init__()
        self.message = message


class Result:
    status: str
    time: int
    memory: int

    def __init__(self, status='', time=-1, memory=-1):
        self.status = status
        self.time = time
        self.memory = memory

    def get_color(self):
        if self.status == 'AC':
            return 'on_green'
        elif self.status == 'WA':
            return 'on_yellow'
        else:
            return 'on_red'


class Executer:
    sanddir: Path  # = curdir / 'sand'
    executer: Popen

    def __init__(self):
        logger.info('Launch executer')
        # make sandbox dir
        self.sanddir = Path.cwd() / 'sand'
        if self.sanddir.exists():
            rmtree(self.sanddir)
        self.sanddir.mkdir()
        self.sanddir.chmod(0o777)

        env = environ.copy()
        env['POSTGRE_PASS'] = 'secret(off course, we should fix this code...)'
        self.executer = Popen(['unshare', '-fpnm', '--mount-proc',
                               './executer.py'], stdin=PIPE, stdout=PIPE, env=env)

    def run(self, execcmd, copyfiles: [Path], stdin: Path = None, stdout: Path = None, timelimit: float = 2.0) -> Result:
        data = {
            'exec': execcmd,
            'timelimit': timelimit,
        }
        if copyfiles:
            data['files'] = list(map(str, copyfiles))
        if stdin:
            data['stdin'] = str(stdin)
        if stdout:
            data['stdout'] = str(stdout)

        logger.info('judge -> executer data: {}'.format(data))
        with open('work/comm.json', 'w') as f:
            f.write(json.dumps(data))
        self.executer.stdin.write(b'comm\n')
        self.executer.stdin.flush()

        s = self.executer.stdout.readline().decode('utf-8').strip()
        if s != 'OK':
            logger.error('Error executer: {}'.format(s))
            return Result()
        logger.info('Return OK')
        result = json.load(open('work/resp.json', 'r'))
        return Result(result['status'], result['time'], result['memory'])


logger.info('Make workdir')
workdir = Path.cwd() / 'work'

if workdir.exists():
    rmtree(workdir)
workdir.mkdir()
shutil.copy('testlib.h', workdir / 'testlib.h')


class Judgement:
    executer: Executer

    def __init__(self):
        self.executer = Executer()

    def compile(self, src: Path, lang: str, copyfiles: [str] = []):
        copyfiles.append(src)
        if lang == 'cpp':
            self.executer.run(
                'g++ -O2 -std=c++14 -o {} {}'.format(src.stem, src.name),
                copyfiles, timelimit=30.0)
            shutil.copy(self.executer.sanddir / src.stem, workdir / src.stem)
        else:
            print('Unknown type of file {}'.format(src))
            raise UnknownTypeFile('Unknown file: {}'.format(src))

    def single(self, inpath: str, outpath: str, timelimit: float = 2.0):
        shutil.copy(inpath, workdir / 'case.in')
        shutil.copy(outpath, workdir / 'case.out')
        inpath = workdir / 'case.in'
        outpath = workdir / 'case.out'
        anspath = workdir / 'case.ans'

        # run
        result = self.executer.run(
            './main', [workdir / 'main'], inpath, anspath, timelimit)

        if result.status == 'OK':
            checker_result = self.executer.run('./checker case.in case.out case.ans',
                                               copyfiles=[workdir / 'checker',
                                                          inpath, outpath, anspath],
                                               timelimit=30.0)
            if checker_result.status == 'OK':
                result.status = 'AC'
            elif checker_result.status == 'RE':
                result.status = 'WA'
            elif checker_result.status == 'TLE':
                result.status = 'ITLE'
            else:
                result.status = 'IE'

        logger.info('judged {} res={} {} msecs'.format(
            inpath, colored(result.status, on_color=result.get_color()), result.time))
        return result

    # assume prepared: work/in, work/out, work/main.cpp, work/checker.cpp

    def judge(self, src: str, lang: str, handler):
        logger.info('Compile checker & source')
        self.compile(workdir / 'checker.cpp', 'cpp', [workdir / 'testlib.h'])
        self.compile(src, lang)

        file_list = list(sorted(workdir.glob('in/*')))
        for i, inpath in enumerate(file_list):
            stem = inpath.stem
            outpath = workdir / 'out' / (stem + '.out')
            result = self.single(inpath=inpath, outpath=outpath)

            handler(stem, result)

        logger.info('End judge')


def fetchdata(conn, problemid):
    # get zip file
    testhash = ''
    with conn.cursor() as cursor:
        if cursor.execute('select testhash from problems where name = %s', (problemid, )) == 0:
            return
        testhash = cursor.fetchone()[0]

    zippath = workdir / 'cases-{}.zip'.format(testhash)

    if not zippath.exists():
        logger.info('Nothing {}, fetching...'.format(zippath))
        with conn.cursor() as cursor:
            if cursor.execute('select testzip from problems where name = %s', (problemid, )) == 0:
                return
            zipdata = cursor.fetchone()[0]
            with open(zippath, 'wb') as f:
                f.write(zipdata)

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


# create table submission_testcase_results(
#     submission int,       -- primary main
#     testcase varchar(32), -- primary sub
#     status varchar(32),
#     maxtime int,
#     maxmemory int,
#     primary key(submission, testcase)
# )


def judge(conn, subid: int):
    logger.info('Judge start submission id = {}'.format(subid))
    with conn.cursor() as cursor:
        cursor.execute('update submissions set status = %s where id = %s',
                        ('Judging', subid))
        conn.commit()

    logger.info('Fetch data from SQL')
    submission = None
    with conn.cursor() as cursor:
        if cursor.execute('select problem, lang, source from submissions where id = %s', (subid, )) == 0:
            return
        submission = cursor.fetchone()

    fetchdata(conn, submission[0])

    # write source
    with open(workdir / 'main.cpp', 'w') as f:
        f.write(submission[2])

    all_result = Result('AC')

    def refresh(name: str, result: Result):
        with conn.cursor() as cursor:
            cursor.execute('''insert into submission_testcase_results
                              (submission, testcase, status, time, memory)
                              values (%s, %s, %s, %s, %s)''',
                           (subid, name, result.status, result.time, result.memory))
            conn.commit()
        if result.status != 'AC':
            all_result.status = result.status
        all_result.time = max(all_result.time, result.time)
        all_result.memory = max(all_result.memory, result.memory)

    judgement = Judgement()
    judgement.judge(workdir / 'main.cpp', 'cpp', refresh)

    with conn.cursor() as cursor:
        cursor.execute('update submissions set status = %s, maxtime = %s, maxmemory = %s where id = %s',
                       (all_result.status, all_result.time, all_result.memory, subid))
        conn.commit()

    logger.info('End judge')


if __name__ == "__main__":
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

        subid = res[1]
        try:
            judge(conn, subid)
        except Exception as e:
            logger.exception(e)
            with conn.cursor() as cursor:
                cursor.execute('update submissions set status = %s where id = %s',
                               ('IE', subid))
                conn.commit()
