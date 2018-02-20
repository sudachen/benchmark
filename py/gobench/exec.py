
import os
import os.path
import tempfile
import subprocess
from . import util
from .util import Success, Fail

class Env(object):
    __slots__ = ('generic', 'windows', 'unix')

    def __init__(self,generic=None,windows=None,unix=None):

        def mapenv(e):
            if e is None:
                return {}
            elif isinstance(e, list):
                return dict(e)
            elif isinstance(e, dict):
                return e
            else:
                raise TypeError("bad environment value type")

        self.generic = mapenv(generic)
        self.windows = mapenv(windows)
        self.unix = mapenv(unix)

    def get(self):
        return {**os.environ,**self.generic}


class ExecutorError(Exception):
    def __init__(self,text):
        super(Exception, self).__init__(self,text)


class UnknownExtension(ExecutorError):
    def __init__(self,ext):
        super(ExecutorError, self).__init__(self,"have no executor for *.{}".format(ext))


class Executor(object):
    __slots__ = ('env', 'temp', 'stderr', 'stdout', 'workdir')

    def __init__(self, workdir, env, temp, stdout_filename=None, stderr_filename=None):
        self.env = env.get()
        self.temp = temp
        self.workdir = workdir
        delete = True if temp is None else False
        if stdout_filename is None:
            self.stdout = tempfile.NamedTemporaryFile(mode="w+t", dir=temp, delete=delete)
        else:
            self.stdout = open(stdout_filename,"w+")
        if stderr_filename is None:
            self.stderr = tempfile.NamedTemporaryFile(mode="w+t", dir=temp, delete=delete)
        else:
            self.stderr = open(stderr_filename,"w+")

    def run(self,*args):
        try:
            gopath = os.getenv("GOPATH",None)
            if gopath is not None and self.workdir.startswith(gopath):
                util.verbose("in the dir $(GOPATH){}", self.workdir[len(gopath):])
            else:
                util.verbose("in the dir {}", self.workdir)
            util.verbose("\texecuting: {}"," ".join(args))
            with subprocess.Popen(args,stdout=self.stdout, stderr=self.stderr, env=self.env, cwd=self.workdir) as p:
                result = p.wait()
            if result == 0:
                return Success
            else:
                return Fail("with exit code {}".format(result))
        except subprocess.SubprocessError as e:
            return Fail("with SubprocessError({})".format(e))
        except OSError as e:
            return Fail("with OSError({})".format(e))
