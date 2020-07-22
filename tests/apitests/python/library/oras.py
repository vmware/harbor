# -*- coding: utf-8 -*-

import os
import base
from datetime import datetime

oras_cmd = "oras"
file_artifact = "artifact.txt"
file_readme = "readme.md"
file_config = "config.json"

def oras_push(harbor_server, user, password, project, repo, tag):
    oras_login(harbor_server, user, password)
    fo = open(file_artifact, "w")
    fo.write( "hello artifact" )
    fo.close()
    md5_artifact = base.run_command( ["md5sum", file_artifact] )
    fo = open(file_readme, "w")
    fo.write( r"Docs on this artifact" )
    fo.close()
    md5_readme = base.run_command( [ "md5sum", file_readme] )
    fo = open(file_config, "w")
    fo.write( "{\"doc\":\"readme.md\"}" )
    fo.close()
    ret = base.run_command( [oras_cmd, "push", harbor_server + "/" + project + "/" + repo+":"+ tag,
                             "--manifest-config", "config.json:application/vnd.acme.rocket.config.v1+json", \
                             file_artifact+":application/vnd.acme.rocket.layer.v1+txt", \
                             file_readme +":application/vnd.acme.rocket.docs.layer.v1+json"] )
    return md5_artifact.split(' ')[0], md5_readme.split(' ')[0]

def oras_login(harbor_server, user, password):
     ret = base.run_command([oras_cmd, "login", "-u", user, "-p", password, harbor_server])

def oras_pull(harbor_server, user, password, project, repo, tag):
    try:
        cwd = os.getcwd()
        cwd= cwd + r"/tmp" + datetime.now().strftime(r'%m%s')
        if os.path.exists(cwd):
          os.rmdir(cwd)
        os.makedirs(cwd)
        os.chdir(cwd)
        print "Tmp dir:", cwd
    except Exception as e:
        raise Exception('Error: Exited with error {}',format(e))
    ret = base.run_command([oras_cmd, "pull", harbor_server + "/" + project + "/" + repo+":"+ tag, "-a"])
    assert os.path.exists(file_artifact)
    assert os.path.exists(file_readme)
    return base.run_command( ["md5sum", file_artifact] ).split(' ')[0], base.run_command( [ "md5sum", file_readme] ).split(' ')[0]
