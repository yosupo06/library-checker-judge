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
from copy import deepcopy
from datetime import datetime
from logging import Logger, basicConfig, getLogger
from os import environ, getenv
from pathlib import Path
from shutil import copy, rmtree
from subprocess import (DEVNULL, PIPE, STDOUT, CalledProcessError, Popen,
                        TimeoutExpired, check_call, run)
from time import sleep

import toml
from termcolor import colored, cprint

basicConfig(
    level=getenv('LOG_LEVEL', 'DEBUG'),
    format="%(asctime)s %(levelname)s %(name)s :%(message)s"
)
logger: Logger = getLogger(__name__)

workdir = Path.cwd() / 'work'


class Lang:
    source: str = ''
    compile: str = ''
    objects: [str] = []
    exec: str = ''

    __langsinfo = None

    def __init__(self, lang: str):
        if not Lang.__langsinfo:
            Lang.__langsinfo = toml.load(
                open('../compiler/langs.toml'))['langs']

        langinfo = Lang.__langsinfo[lang]

        self.source = langinfo['source']
        self.compile = langinfo['compile']
        self.objects = langinfo['objects']
        self.exec = langinfo['exec']


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

    def clean(self):
        self.executer.stdin.write(b'clean\n')
        self.executer.stdin.flush()
        s = self.executer.stdout.readline().decode('utf-8').strip()
        if s != 'OK':
            logger.error('Error executer clean: {}'.format(s))

    def run(self,
            exec: str,
            timelimit: float,
            sendfiles: [str] = [],
            getfiles: [str] = [],
            stdin: Path = None,
            stdout: Path = None) -> Result:

        self.clean()
        for f in sendfiles:
            shutil.copy(workdir / f, self.sanddir / f)

        data = {
            'exec': exec,
            'timelimit': timelimit,
        }
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
        result = json.load(open('work/resp.json', 'r'))
        logger.info('Return OK status: {}'.format(result['status']))

        result = Result(result['status'], result['time'], result['memory'])

        for f in getfiles:
            if not (self.sanddir / f).exists() or not (self.sanddir / f).is_file():
                result.status = 'RE'
                break
            shutil.copy(self.sanddir / f, workdir / f)

        return result

    def kill(self):
        self.executer.stdin.write(b'last\n')
        self.executer.stdin.flush()
        self.executer.wait()


class Judgement:
    executer: Executer
    lang: str

    def __init__(self, langname: str):
        self.executer = Executer()
        self.lang = Lang(langname)

    def single(self, inpath: str, outpath: str, timelimit: float):
        shutil.copy(inpath, workdir / 'case.in')
        shutil.copy(outpath, workdir / 'case.out')

        result = self.executer.run(
            exec=self.lang.exec,
            sendfiles=self.lang.objects,
            stdin=workdir / 'case.in',
            stdout=workdir / 'case.your',
            timelimit=timelimit)
        if result.status == 'OK':
            lang_checker = Lang('checker')
            exec_command = lang_checker.exec.format(
                input='case.in',
                judge='case.out',
                contestant='case.your',
            )
            objects = deepcopy(lang_checker.objects)
            objects.extend(['case.in', 'case.out', 'case.your'])
            checker_result = self.executer.run(
                exec=exec_command,
                sendfiles=objects,
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
    def compile_checker(self) -> bool:
        lang_checker = Lang('checker')
        result = self.executer.run(
            exec=lang_checker.compile,
            sendfiles=['checker.cpp', 'testlib.h'],
            getfiles=lang_checker.objects,
            timelimit=30.0
        )
        return result.status == 'OK'

    def compile(self) -> bool:
        result = self.executer.run(
            exec=self.lang.compile,
            sendfiles=[self.lang.source],
            getfiles=self.lang.objects,
            timelimit=30.0
        )
        return result.status == 'OK'

    def judge(self, handler, timelimit):
        file_list = list(sorted(workdir.glob('in/*')))
        for i, inpath in enumerate(file_list):
            stem = inpath.stem
            outpath = workdir / 'out' / (stem + '.out')
            result = self.single(inpath=inpath, outpath=outpath, timelimit=timelimit)

            handler(stem, result)

        self.executer.kill()
        logger.info('End judge')

