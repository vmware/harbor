# -*- coding: utf-8 -*-

import base
import json
import docker_api

def load_bundle(service_image, invocation_image):
    bundle_file = "./tests/apitests/python/bundle_data/bundle.json"
    bundle_tmpl_file = "./tests/apitests/python/bundle_data/bundle.json.tmpl"
    with open(bundle_tmpl_file,'r') as load_f:
        load_dict = json.load(load_f)
        print "load_dict:", load_dict
        print "load_dict-invocationImages:", load_dict["invocationImages"][0]["contentDigest"]
        load_dict["images"]["hello"]["image"] = service_image
        load_dict["invocationImages"][0]["image"] = invocation_image
        bundle_str = json.dumps(load_dict)
        print "bundle_str:", bundle_str
        with open(bundle_file,'w') as dump_f:
            dump_f.write(bundle_str)
            dump_f.close()
        return bundle_file

def cnab_fixup_bundle(bundle_file, target, auto_update_bundle = True):
    fixed_bundle_file = "./tests/apitests/python/bundle_data/fixed-bundle.json"
    command = ["sudo", "cnab-to-oci", "--log-level", "debug", "fixup", bundle_file, "--target", target, "--bundle", fixed_bundle_file]
    if auto_update_bundle == True:
         command.append("--auto-update-bundle")
         #fixed_bundle_file = bundle_file
    print "Command: ", command
    ret = base.run_command(command)
    print "Command return: ", ret
    return fixed_bundle_file

def cnab_push_bundle(bundle_file, target):
    command = ["cnab-to-oci", "push", bundle_file, "--target", target, "--auto-update-bundle"]
    print "Command: ", command
    ret = base.run_command(command)
    print "Command return: ", ret
    for line in ret.split("\n"):
        line = line.replace('\"', '')
        if line.find('sha256') >= 0:
            return line[-71:]
    raise Exception(r"Fail to get sha256 in returned data: {}".format(ret))

def push_cnab_bundle(harbor_server, user, password, service_image, invocation_image, target, auto_update_bundle = True):
    docker_api.docker_info_display()
    docker_api.docker_login_cmd(harbor_server, user, password, enable_manifest = False)
    bundle_file = load_bundle(service_image, invocation_image)
    fixed_bundle_file = cnab_fixup_bundle(bundle_file, target, auto_update_bundle = auto_update_bundle)
    sha256 = cnab_push_bundle(fixed_bundle_file, target)
    print "sha256:", sha256
    return sha256
