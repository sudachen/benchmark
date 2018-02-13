
import os
import os.path
import json
from . import exec

BENCHMARK_FILE = 'benchmark.go'


class Benchmark(object):
    __slots__ = ('label', 'env')

    def __init__(self, label, env=exec.Env()):
        self.label = label
        self.env = env

    def execute(self, the_branch, temp=None, pprof=None, callgraph=None):
        workdir = the_branch.dirname(self.label)
        ex = exec.Executor(workdir, self.env, temp)

        if pprof is True:
            ppfopt = ['--pprof']
        elif isinstance(pprof,str):
            ppfopt = ['--pprof','--cpuprof='+pprof]
        else:
            ppfopt = ()

        if callgraph is True:
            pngopt = ['--callgraph=0']
        elif isinstance(callgraph,int):
            pngopt = ['--callgraph='+str(callgraph)]
        else:
            pngopt = ()


        status = ex.run("go","run",BENCHMARK_FILE,*ppfopt,*pngopt)
        ex.stdout.seek(0)
        ex.stderr.seek(0)

        if status is not exec.Success:
            raise ExecutionBenchmarkError(self.label, status.reason)

        return load(the_branch,ex.stdout)


class BenchmarkError(Exception):
    def __init__(self,text):
        super(Exception, self).__init__(self, text)


class UnknownBenchmarkError(BenchmarkError):
    def __init__(self,benchmark_label):
        super(BenchmarkError, self).__init__(self, "unknown benchmark {}".format(benchmark_label))


class ExecutionBenchmarkError(BenchmarkError):
    def __init__(self,benchmark_label,reason):
        super(BenchmarkError, self).__init__(self, "benchmark {} failed: {}".format(benchmark_label,reason))


class MsgKind(object):
    def __str__(self):
        if self is MsgError:
            return "Error"
        if self is MsgInfo:
            return "Info"
        if self is MsgDebug:
            return "Debug"
        if self is MsgOpt:
            return "Opt"
        raise ValueError()


MsgError = MsgKind()
MsgDebug = MsgKind()
MsgInfo = MsgKind()
MsgOpt = MsgKind()


class Message(object):
    __slots__ = ('kind', 'text')

    def __init__(self,kind,text):
        self.kind = kind
        self.text = text

    def __repr__(self):
        return 'Message(kind="{}", text="{}")'.format(
            self.kind,
            self.text
        )


class Task(object):
    __slots__ = ('label', 'total', 'active', 'count', 'error', 'children', 'messages')

    def __init__(self, label, total, active, count, error, children, messages):
        self.label = label
        self.total = total
        self.active = active
        self.count = count
        self.error = error
        self.children = children
        self.messages = messages

    def __repr__(self):
        return 'Task(label="{}", total={}, active={}, count={}, error={}, children={}, messages={})'.format(
            self.label,
            self.total,
            self.active,
            self.count,
            repr(self.error),
            self.children,
            self.messages
        )


class PprofRow(object):
    __slots__ = ('flat','flatP','sumP','cum','cumP','function')
    columns = ("flat","flat%","sum%","cum","cum%","function")

    def __init__(self, flat, flatP, sumP, cum, cumP, function):
        self.flat = flat
        self.flatP = flatP
        self.sumP = sumP
        self.cum = cum
        self.cumP = cumP
        self.function = function

    def __repr__(self):
        return "PprofRow(flat={}, flatp={}, sumP={}, cum={}, cumP={} function='{}')".format(
            self.flat,self.flatP,self.sumP,self.cum,self.cumP,self.function)

    def __getitem__(self, item):
        if item == "flat":
            return self.flat
        if item == "flat%":
            return self.flatP
        if item == "sum%":
            return self.sumP
        if item == "cum":
            return self.cum
        if item == "cum%":
            return self.cumP
        if function == "function":
            return self.cumP
        raise KeyError(item)

    def __iter__(self):
        yield self.flat
        yield self.flatP
        yield self.sumP
        yield self.cum
        yield self.cumP
        yield self.function

    def __len__(self):
        return len(self.columns)

class PprofUnit(object):
    __slots__ = ['label']

    def __init__(self, label):
        self.label = label

    def __repr__(self):
        return self.label


Msec = PprofUnit("ms")
Usec = PprofUnit("us")
Sec = PprofUnit("s")


class Pprof(object):
    __slots__ = ('label', 'unit', 'rows', 'errors', 'image')

    def __init__(self, label, unit, rows, errors, image):
        self.label = label
        self.image = image
        if unit == 'ms':
            self.unit = Msec
        elif unit == 'us':
            self.unit = Usec
        elif unit == 's':
            self.unit = Sec
        self.rows = rows
        self.errors = errors

    def __repr__(self):
        return "Pprof(label='{}', unit='{}', rows={}, errors={})".format(
            self.label, self.unit, self.rows, self.errors)


def load_results(f):

    def decode_object(m):
        if "kind" in m:
            kind = m["kind"]
            if kind == "MsgError":
                kind = MsgError
            elif kind == "MsgInfo":
                kind = MsgInfo
            elif kind == "MsgDebug":
                kind = MsgDebug
            elif kind == "MsgOpt":
                kind = MsgOpt
            else:
                raise ValueError()
            return Message(kind,m["text"])
        elif "flat%" in m:
            return PprofRow(
                float(m["flat"]),
                float(m["flat%"]),
                float(m["sum%"]),
                float(m["cum"]),
                float(m["cum%"]),
                m["function"]
            )
        elif "rows" in m:
            return Pprof(
                m["label"],
                m["unit"],
                m.get("rows",None),
                m.get("errors",None),
                m.get("image","")
            )
        elif "label" in m:
            t = Task(
                m["label"],
                int(m["total"]),
                int(m["active"]),
                int(m["count"]),
                m.get("error",None),
                m.get("children",None),
                m.get("messages",None),
            )
            if m["label"] == '.':
                return (t,m.get("pprof",None))
            return t
        return m

    return json.load(f,object_hook=decode_object)


class Result(object):
    __slots__ = ('branch', 'results', 'pprof')

    def __init__(self, branch, results, pprof):
        self.branch = branch
        self.results = results
        self.pprof = pprof

    def __repr__(self):
        return "Result(branch='{}', results={}, pprof={})".format(
            self.branch,
            self.results,
            self.pprof
        )


def load(branch,file):
    r, ppf = load_results(file)
    return Result(branch, r, {i.label:i for i in ppf} )

