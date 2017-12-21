#!/usr/bin/python2

import os, subprocess
import time

SHELL_SCRIPT_DIR = '/harbor/workspace/harbor_nightly_test/tests/nightly-test/shellscript/'

def getvmip(vc_url, vc_user, vc_password, vm_name, timeout=600) :
    cmd = (SHELL_SCRIPT_DIR+'getvmip.sh %s %s %s %s ' % (vc_url, vc_user, getPasswordInShell(vc_password), vm_name))
    print cmd
    interval = 10
    while True:
        try:
            if timeout <= 0:
                print "timeout to get ova ip"
                return -1
            result = subprocess.check_output(cmd,shell=True)
            print "######"
            print result
            print result == ''
            print "######"            
            if result is not '' and result is not "photon-machine":
                print result
                return 0
        except Exception, e:
            timeout -= interval
            time.sleep(interval)
            continue
        timeout -= interval
        time.sleep(interval)
    return result

def destroyvm(vc_url, vc_user, vc_password, vm_name) :
    cmd = (SHELL_SCRIPT_DIR+'destroyvm.sh %s %s %s %s ' % (vc_url, vc_user, getPasswordInShell(vc_password), vm_name))  
    result = subprocess.check_output(cmd, shell=True)
    return result

def getPasswordInShell(password) :
    return password.replace("!", "\!")
