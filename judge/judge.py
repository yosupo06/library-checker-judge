#!/usr/bin/env python3

import os, argparse
import psycopg2
import shutil
import glob
import tempfile
from time import sleep
from subprocess import run, check_call, TimeoutExpired, CalledProcessError, DEVNULL
from termcolor import colored, cprint
from datetime import datetime

print('[*] launch judge.py')

curdir = os.path.abspath(os.path.curdir)

# connect sql
# TODO: .pginit
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
def judgecase(execcmd, inpath, anspath, timelimit = 2.0):
    check_call(['cp', inpath, 'sand/in.txt'])

    # run
    result = run_in_sandbox(execcmd, stdin = open(inpath, 'r'), stdout = open('work/out.txt', 'w'))
    color = ''
    if result.result == 'OK':
        try:
            # output check
            check_call(['diff', 'work/out.txt', anspath], stdout=DEVNULL)
        except CalledProcessError:
            result.result = 'WA'
            color = 'on_yellow'
        else:
            result.result = 'AC'
            color = 'on_green'
    else:
        color = 'on_red'

    print('[*] judged {} res={} {} msecs'.format(inpath, colored(result.result, on_color=color), result.time))
    return result

# source must be abspath
def compilecxx(srcpath):
    check_call(['cp', srcpath, 'sand/main.cpp'])
    run_in_sandbox('g++ -O2 -std=c++14 -o main main.cpp')

def judge(subid):    
    print('[!] judge start id = {}'.format(subid))
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

    with open('work/main.cpp', 'w') as f:
        f.write(submittion[2])
    
    with open('work/cases.zip', 'wb') as f:
        f.write(problem[0])

    # unzip    
    if os.path.exists('in'):
        shutil.rmtree('in')

    if os.path.exists('out'):
        shutil.rmtree('out')
    check_call(['unzip', 'work/cases.zip'])


    print('[!] end fetch data')

    print('[*] compile main.cpp')
    compilecxx(os.path.abspath('work/main.cpp'))

    status = 'AC'
    consume_time = -1
    consume_memory = -1
    file_list = list(sorted(glob.glob(curdir + '/in/*')))
    for i, inpath in enumerate(file_list):
        _, filepath = os.path.split(inpath)
        name, _ = os.path.splitext(filepath)
        result = judgecase('./main', inpath, curdir + '/out/' + name + '.out')

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

    print('[!] end judge')

# read sql queue and judge
while True:
    sleep(1)
    sql = 'select id from queue'
    res = None
    with conn.cursor() as cursor:
        cursor.execute('select id, submittion from tasks limit 1')
        res = cursor.fetchone()
        if res == None:
            print('waiting...')
            continue            
        cursor.execute('delete from tasks where id = %s', (res[0],))
        conn.commit()

    judge(res[1])


conn.close()
