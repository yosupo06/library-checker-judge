#!/usr/bin/env python3

import os, argparse
import psycopg2
import shutil
import glob
import zipfile
import tempfile
from time import sleep
from subprocess import run, check_call, TimeoutExpired, CalledProcessError, DEVNULL
from termcolor import colored, cprint
from datetime import datetime
from logging import getLogger, basicConfig

basicConfig(
    level = os.getenv('LOG_LEVEL', 'DEBUG'),
    format = "%(asctime)s %(levelname)s %(name)s :%(message)s"
)
logger = getLogger(__name__)

curdir = os.path.abspath(os.path.curdir)
sanddir = os.path.join(curdir, 'sand')
workdir = os.path.join(curdir, 'work')

logger.info('Launch judge.py')


logger.info('Connect SQL')
hostname = os.environ.get('POSTGRE_HOST', '127.0.0.1')
port = int(os.environ.get('POSTGRE_PORT', '5432'))
user = os.environ.get('POSTGRE_USER', 'postgres')
password = os.environ.get('POSTGRE_PASS', 'passwd')
conn = psycopg2.connect(
    host=hostname,
    port=port,
    user=user,
    password=password,
    database='librarychecker'
)

logger.info('Run prepare.sh')
run(['./prepare.sh'])



class Result:
    result = ''
    time = 0
    memory = 0
    def __init__(self, result, time, memory):
        self.result = result
        self.time = time
        self.memory = memory


def run_in_sandbox(execcmd, stdin = None, stdout = None, timelimit = 2.0):
    memory_max_usage = '/sys/fs/cgroup/memory/lib-judge/memory.max_usage_in_bytes'
    with open(memory_max_usage, 'w') as f:
        f.write('0')

    result = Result('IE', -1, -1)
    start = datetime.now()
    try:
        check_call(['./exec.sh', execcmd], stdin = stdin, stdout = stdout, timeout=timelimit)
    except TimeoutExpired:
        result.result = 'TLE'
    except CalledProcessError:
        result.result = 'RE'
    else:
        end = datetime.now()
        result.result = 'OK'
        result.time = (end - start).seconds * 1000 + (end - start).microseconds // 1000
        with open(memory_max_usage, 'r') as f:
            result.memory = int(f.read())

    return result

# TODO: output_checker
def judgecase(execcmd, inpath, outpath, timelimit = 2.0):
    anspath = os.path.join(workdir, 'ans.txt')
    shutil.copy(inpath, os.path.join(sanddir, 'in.txt'))

    # run
    result = run_in_sandbox(execcmd, stdin = open(inpath, 'r'), stdout = open(anspath, 'w'))
    color = ''
    if result.result == 'OK':
        try:
            # output check
            check_call(['diff', anspath, outpath], stdout=DEVNULL)
        except CalledProcessError:
            result.result = 'WA'
            color = 'on_yellow'
        else:
            result.result = 'AC'
            color = 'on_green'
    else:
        color = 'on_red'

    logger.info('judged {} res={} {} msecs'.format(inpath, colored(result.result, on_color=color), result.time))
    return result

# source must be abspath
def compilecxx(srcpath):
    check_call(['cp', srcpath, 'sand/main.cpp'])
    run_in_sandbox('g++ -O2 -std=c++14 -o main main.cpp')

def judge(subid):    
    logger.info('Judge start submittion id = {}'.format(subid))

    logger.info('Fetch data from SQL')
    submittion = None
    problem = None
    with conn.cursor() as cursor:
        if cursor.execute('select problem, lang, source from submittions where id = %s', (subid, )) == 0:
            return        
        submittion = cursor.fetchone()

    with conn.cursor() as cursor:
        if cursor.execute('select testzip source from problems where name = %s', (submittion[0], )) == 0:
            return
        problem = cursor.fetchone()

    logger.info('Extact fetched data')
    srcpath = os.path.join(workdir, 'main.cpp') # Todo: other lang
    zippath = os.path.join(workdir, 'cases.zip')
    indir = os.path.join(workdir, 'in')
    outdir = os.path.join(workdir, 'out')
    with open(os.path.join(workdir, 'main.cpp'), 'w') as f:
        f.write(submittion[2])
    
    with open(zippath, 'wb') as f:
        f.write(problem[0])

    if os.path.exists(indir):
        shutil.rmtree(indir)

    if os.path.exists(outdir):
        shutil.rmtree(outdir)
    
    with zipfile.ZipFile(zippath, 'r') as f:
        f.extractall(workdir)


    logger.info('Compile main.cpp')
    compilecxx(srcpath)

    status = 'AC'
    consume_time = -1
    consume_memory = -1
    file_list = list(sorted(glob.glob(indir + '/*')))
    for i, inpath in enumerate(file_list):
        _, filepath = os.path.split(inpath)
        name, _ = os.path.splitext(filepath)
        result = judgecase('./main', inpath, os.path.join(outdir, name + '.out'))

        with conn.cursor() as cursor:
            cursor.execute('update submittions set status = %s where id = %s',
                ('{}/{}'.format(i, len(file_list)), subid))
            conn.commit()

        if result.result != 'AC':
            status = result.result
        consume_time = max(consume_time, result.time)
        consume_memory = max(consume_memory, result.memory)

    with conn.cursor() as cursor:
        cursor.execute('update submittions set status = %s, maxtime = %s, maxmemory = %s where id = %s',
            (status, consume_time, consume_memory, subid))
        conn.commit()

    logger.info('End judge')

# read sql queue and judge
while True:
    sleep(1)
    sql = 'select id from queue'
    res = None
    with conn.cursor() as cursor:
        cursor.execute('select id, submittion from tasks limit 1')
        res = cursor.fetchone()
        if res == None:
            continue            
        cursor.execute('delete from tasks where id = %s', (res[0],))
        conn.commit()

    judge(res[1])


conn.close()
