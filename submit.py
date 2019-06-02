#!/usr/bin/env python3

import os, argparse
import psycopg2
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

print('[*] connect SQL')
# connect sql
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

sql = 'insert into submittions(submittime, problem, lang, source, status) values (%s, %s, %s, %s, %s)'
with conn.cursor() as cursor:
    cursor.execute(sql, (datetime.datetime.now(), problem, ext, source, 'WJ'))
    cursor.execute("select currval('submittions_id_seq')")
    id = cursor.fetchone()[0]
    cursor.execute('insert into tasks(submittion) values (%s)', (id, ))
    conn.commit()
conn.close()
