
import os.path
import subprocess
from tempfile import NamedTemporaryFile
from . import util
from . import bench
from .util import Return, Fail

BENCHMARK_PPROF = "benchmark.pprof"

def run(workdir,*args):
    try:
        with NamedTemporaryFile(mode="w+b") as o:
            util.verbose("in the dir {}".format(workdir))
            util.verbose("\texecuting: {}"," ".join(args))
            with subprocess.Popen(args,stdout=o,cwd=workdir) as p:
                result = p.wait()
            if result == 0:
                o.seek(0)
                return Return(o.read())
            else:
                return Fail("with exit code {}".format(result))
    except subprocess.SubprocessError as e:
        return Fail("with SubprocessError({})".format(e))
    except OSError as e:
        return Fail("with OSError({})".format(e))

def pprof(pprof, *args):
    work_dir=os.path.join(os.environ["GOPATH"],"src","github.com","sudachen","benchmark","cmd","pprof")
    cmd = ["go","run","pprof.go"]
    cmd.extend(args)
    cmd.append(pprof)
    return run(work_dir,*cmd)
