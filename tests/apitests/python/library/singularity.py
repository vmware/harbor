# -*- coding: utf-8 -*-

import os
import base
from datetime import datetime

singularity_cmd = "singularity"
timestamp = datetime.now().strftime(r'%m%s')

def set_singularity_login_env(user, password):
    os.environ.setdefault('SINGULARITY_DOCKER_USERNAME', user)
    os.environ.setdefault('SINGULARITY_DOCKER_PASSWORD', password)

def singularity_push_to_harbor(harbor_server, sif_file, project, image, tag):
    ret = base.run_command( [singularity_cmd, "push", sif_file, "oras://"+harbor_server + "/" + project + "/" + image+":"+ tag] )

def singularity_pull(out_file, from_sif_file):
    ret = base.run_command( [singularity_cmd, "pull", out_file, from_sif_file] )

def push_singularity_to_harbor(from_URI, from_namespace, harbor_server, user, password, project, image, tag):
    tmp_sif_file = image+timestamp+".sif"
    set_singularity_login_env(user, password)
    singularity_pull(tmp_sif_file, from_URI+"//"+from_namespace + image+":" + tag)
    singularity_push_to_harbor(harbor_server, tmp_sif_file, project, image, tag)
