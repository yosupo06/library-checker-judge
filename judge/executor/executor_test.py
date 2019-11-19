#!/usr/bin/env python3

import json
import unittest
from logging import basicConfig, getLogger
from os import getenv
from pathlib import Path
from shutil import copy
from subprocess import PIPE, run
from tempfile import NamedTemporaryFile, TemporaryDirectory
from uuid import uuid4

executor = Path('./executor.py').absolute()
logger = getLogger(__name__)


def get_tmpdir(src: Path):
    tmpdir = TemporaryDirectory()
    Path(tmpdir.name).chmod(0o777)
    copy(src, Path(tmpdir.name) / src.name)
    return tmpdir


def get_result(execcmd, cwd, overlay, tl=None, stdin=None, stderr=None):
    logger.info('execute {}'.format(execcmd))
    with NamedTemporaryFile() as resfile:
        cmd = [executor, '--result', resfile.name]
        if overlay:
            cmd = cmd + ['--overlay']
        if tl:
            cmd = cmd + ['--tl', str(tl)]
        if stdin:
            cmd = cmd + ['--stdin', str(stdin)]
        if stderr:
            cmd = cmd + ['--stderr', str(stderr)]
        cmd = cmd + ['--'] + execcmd
        returncode = run(cmd, cwd=cwd).returncode

        result = json.load(resfile)
        logger.info('result {}'.format(result))
        return returncode, result


class TestHelloWorld(unittest.TestCase):
    def test_cpp(self):
        tmpdir = get_tmpdir(Path('../test_src/Hello.cpp'))
        code, result = get_result(['g++', 'Hello.cpp'], tmpdir.name, False)
        self.assertEqual(code, 0)
        self.assertEqual(result['returncode'], 0)
        code, result = get_result(['./a.out'], tmpdir.name, True)
        self.assertEqual(code, 0)
        self.assertEqual(result['returncode'], 0)
        self.assertLess(result['time'], 0.05)
        tmpdir.cleanup()

    def test_cpp_with_flag(self):
        tmpdir = get_tmpdir(Path('../test_src/Hello.cpp'))
        code, result = get_result(
            ['g++', 'Hello.cpp', '-o', 'Hello'], tmpdir.name, False)
        self.assertEqual(code, 0)
        self.assertEqual(result['returncode'], 0)
        code, result = get_result(['./Hello'], tmpdir.name, True)
        self.assertEqual(code, 0)
        self.assertEqual(result['returncode'], 0)
        self.assertLess(result['time'], 0.05)
        tmpdir.cleanup()

    def test_java(self):
        tmpdir = get_tmpdir(Path('../test_src/Hello.java'))
        code, result = get_result(['javac', 'Hello.java'], tmpdir.name, False)
        self.assertEqual(code, 0)
        self.assertEqual(result['returncode'], 0)
        code, result = get_result(['java', 'Hello'], tmpdir.name, True)
        self.assertEqual(code, 0)
        self.assertEqual(result['returncode'], 0)
        tmpdir.cleanup()

    def test_java_twoclasses(self):
        tmpdir = get_tmpdir(Path('../test_src/TwoClasses.java'))
        code, result = get_result(
            ['javac', 'TwoClasses.java'], tmpdir.name, False)
        self.assertEqual(code, 0)
        self.assertEqual(result['returncode'], 0)
        code, result = get_result(['java', 'TwoClasses'], tmpdir.name, True)
        self.assertEqual(code, 0)
        self.assertEqual(result['returncode'], 0)
        tmpdir.cleanup()

    def test_python(self):
        tmpdir = get_tmpdir(Path('../test_src/Hello.py'))
        code, result = get_result(['python3', 'Hello.py'], tmpdir.name, True)
        self.assertEqual(code, 0)
        self.assertEqual(result['returncode'], 0)
        tmpdir.cleanup()


class TestTLE(unittest.TestCase):
    def test_tle(self):
        tmpdir = get_tmpdir(Path('../test_src/TLE.cpp'))
        code, result = get_result(['g++', 'TLE.cpp'], tmpdir.name, False)
        self.assertEqual(code, 0)
        self.assertEqual(result['returncode'], 0)
        code, result = get_result(['./a.out'], tmpdir.name, True, 2.0)
        self.assertEqual(code, 124)
        self.assertAlmostEqual(result['time'], 2.0)
        tmpdir.cleanup()


class TestOverlay(unittest.TestCase):
    def test_overlay_false(self):
        tmpdir = TemporaryDirectory()
        Path(tmpdir.name).chmod(0o777)
        code, result = get_result(['touch', 'Hidden'], tmpdir.name, False)
        self.assertEqual(code, 0)
        self.assertEqual(result['returncode'], 0)
        self.assertTrue((Path(tmpdir.name) / 'Hidden').exists())
        tmpdir.cleanup()

    def test_overlay_true(self):
        tmpdir = TemporaryDirectory()
        code, result = get_result(['touch', 'Hidden'], tmpdir.name, True)
        self.assertEqual(code, 0)
        self.assertEqual(result['returncode'], 0)
        self.assertFalse((Path(tmpdir.name) / 'Hidden').exists())
        tmpdir.cleanup()


class TestStdin(unittest.TestCase):
    def test_stdin(self):
        tmpfile = NamedTemporaryFile(mode='w', delete=False)
        tmpfile.write('Test str\n')
        tmpfile.close()
        code, result = get_result(['cat'], '.', False, 2.0, tmpfile.name)
        self.assertEqual(code, 0)
        self.assertEqual(result['returncode'], 0)


class TestTmpDir(unittest.TestCase):
    def test_tmpdir(self):
        tmpdir = TemporaryDirectory()
        name = str(uuid4())
        code, result = get_result(['touch', '/tmp/' + name], tmpdir.name, True)
        self.assertEqual(code, 0)
        self.assertEqual(result['returncode'], 0)
        self.assertFalse((Path('/tmp') / name).exists())
        tmpdir.cleanup()


class TestStackOverFlow(unittest.TestCase):
    def test_tmpdir(self):
        tmpdir = get_tmpdir(Path('../test_src/stack.cpp'))
        code, result = get_result(['g++', 'stack.cpp'], tmpdir.name, False)
        self.assertEqual(code, 0)
        self.assertEqual(result['returncode'], 0)
        code, result = get_result(['./a.out'], tmpdir.name, True)
        self.assertEqual(code, 0)
        self.assertEqual(result['returncode'], 0)
        tmpdir.cleanup()


class TestRE(unittest.TestCase):
    def test_re(self):
        tmpdir = TemporaryDirectory()
        Path(tmpdir.name).chmod(0o777)
        code, result = get_result(['cat', 'dummy'], tmpdir.name, False)
        self.assertEqual(code, 0)
        self.assertNotEqual(result['returncode'], 0)
        tmpdir.cleanup()


class TestForkBomb(unittest.TestCase):
    def test_forkbomb(self):
        logger.info('Start Fork Bomb')
        code, result = get_result(
            ['../test_src/fork_bomb.sh'], '.', True, 10.0, None, "/dev/null")
        logger.info('End')

if __name__ == "__main__":
    basicConfig(
        level=getenv('LOG_LEVEL', 'DEBUG'),
        format="%(asctime)s %(levelname)s %(name)s : %(message)s"
    )
    unittest.main()
