# -*- coding: utf-8 -*-
import subprocess
from testutils import notary_url

def sign_image(registry_ip, project_name, image, tag):
    try:
        ret = subprocess.check_output(["./tests/apitests/python/sign_image.sh", registry_ip, project_name, image, tag, notary_url], shell=False)
        print "sign_image return: ", ret
    except subprocess.CalledProcessError, exc:
        raise Exception("Failed to sign image error is {} {}.".format(exc.returncode, exc.output))

