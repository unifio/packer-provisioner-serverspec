from fabric.api import *

@task
def test():
    run("echo \"This is a test\" >> test.txt")
