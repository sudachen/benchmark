
import os.path

ALL = type("",(),{})

_analisis_dir = None
_root_dir = None

def set_analysis_dir(path):
    global _analysis_dir 
    _analysis_dir = path

def make_temp_dir_with(file=None):
    if file is None:
        dirname = os.path.dirname(os.path.dirname(__file__))
    else:
        dirname = analysis_dir()
    dirname = os.path.join(dirname,".temp")
    if not os.path.exists(dirname):
        os.mkdir(dirname)
    return dirname


def analysis_dir():
    global _analysis_dir
    if _analysis_dir is None:
        dir = os.path.dirname(__file__)
        dir = os.path.dirname(dir)
        _analysis_dir = dir
    return _analysis_dir


def set_root_dir(path):
    global _root_dir
    _root_dir = path

def root_dir():
    global _root_dir
    if _root_dir is None:
        dir =  os.path.dirname(analysis_dir())
        _root_dir = dir
    return _root_dir

def verbose(fmt,*args):
    print(fmt.format(*args))

class SuccessObject(object):
    pass

Success = SuccessObject()

class Return(object):
    __slots__ = ["value"]

    def __init__(self,value):
        self.value = value

class Fail(object):
    __slots__ = ["reason"]

    def __init__(self,reason):
        self.reason = reason

    def __repr__(self):
        return "Fail(reason='{}')".format(self.reason)
