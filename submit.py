#!/usr/bin/env python3

import os, argparse
import pymysql
import datetime

parser = argparse.ArgumentParser(description='Testcase Generator')
parser.add_argument('problem', help='Problem')
parser.add_argument('source', help='Source File')
args = parser.parse_args()

problem = args.problem
sourcepath = args.source
source = ''
with open(sourcepath, 'r') as f:
    source = f.read()
ext = os.path.splitext(sourcepath)[1][1:]

print(problem, sourcepath, ext)

# connect sql
hostname = os.environ.get('MYSQL_HOST', '127.0.0.1')
port = int(os.environ.get('MYSQL_PORT', '3306'))
user = os.environ.get('MYSQL_USER', 'root')
password = os.environ.get('MYSQL_PASS', 'passwd')

conn = pymysql.connect(
    host=hostname,
    port=port,
    user=user,
    password=password,
    database='librarychecker'
)

sql = 'insert into submittion(submittime, problem, lang, source, status) values (%s, %s, %s, %s, %s)'
with conn.cursor() as cursor:
    cursor.execute(sql, (datetime.datetime.now(), problem, ext, source, 'WJ'))
    id = cursor.lastrowid
    cursor.execute('insert into queue(submittion) values (%s)', (id, ))
    conn.commit()
conn.close()
