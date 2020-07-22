# Get or generate private key
import os, subprocess, shutil
from pathlib import Path
from subprocess import DEVNULL
import logging

from g import DEFAULT_GID, DEFAULT_UID, shared_cert_dir, storage_ca_bundle_filename, internal_tls_dir, internal_ca_filename
from .misc import (
    mark_file,
    generate_random_string,
    check_permission,
    stat_decorator,
    get_realpath)

SSL_CERT_PATH = os.path.join("/etc/cert", "server.crt")
SSL_CERT_KEY_PATH = os.path.join("/etc/cert", "server.key")

def _get_secret(folder, filename, length=16):
    key_file = os.path.join(folder, filename)
    if os.path.isfile(key_file):
        with open(key_file, 'r') as f:
            key = f.read()
            print("loaded secret from file: %s" % key_file)
        mark_file(key_file)
        return key
    if not os.path.isdir(folder):
        os.makedirs(folder)
    key = generate_random_string(length)
    with open(key_file, 'w') as f:
        f.write(key)
        print("Generated and saved secret to file: %s" % key_file)
    mark_file(key_file)
    return key


def get_secret_key(path):
    secret_key = _get_secret(path, "secretkey")
    if len(secret_key) != 16:
        raise Exception("secret key's length has to be 16 chars, current length: %d" % len(secret_key))
    return secret_key


def get_alias(path):
    alias = _get_secret(path, "defaultalias", length=8)
    return alias

@stat_decorator
def create_root_cert(subj, key_path="./k.key", cert_path="./cert.crt"):
   rc = subprocess.call(["/usr/bin/openssl", "genrsa", "-out", key_path, "4096"], stdout=DEVNULL, stderr=subprocess.STDOUT)
   if rc != 0:
        return rc
   return subprocess.call(["/usr/bin/openssl", "req", "-new", "-x509", "-key", key_path,\
        "-out", cert_path, "-days", "3650", "-subj", subj], stdout=DEVNULL, stderr=subprocess.STDOUT)

@stat_decorator
def create_cert(subj, ca_key, ca_cert, key_path="./k.key", cert_path="./cert.crt"):
    cert_dir = os.path.dirname(cert_path)
    csr_path = os.path.join(cert_dir, "tmp.csr")
    rc = subprocess.call(["/usr/bin/openssl", "req", "-newkey", "rsa:4096", "-nodes","-sha256","-keyout", key_path,\
        "-out", csr_path, "-subj", subj], stdout=DEVNULL, stderr=subprocess.STDOUT)
    if rc != 0:
        return rc
    return subprocess.call(["/usr/bin/openssl", "x509", "-req", "-days", "3650", "-in", csr_path, "-CA", \
        ca_cert, "-CAkey", ca_key, "-CAcreateserial", "-out", cert_path], stdout=DEVNULL, stderr=subprocess.STDOUT)


def openssl_installed():
    shell_stat = subprocess.check_call(["/usr/bin/which", "openssl"], stdout=DEVNULL, stderr=subprocess.STDOUT)
    if shell_stat != 0:
        print("Cannot find openssl installed in this computer\nUse default SSL certificate file")
        return False
    return True


def prepare_registry_ca(
    private_key_pem_path: Path,
    root_crt_path: Path,
    old_private_key_pem_path: Path,
    old_crt_path: Path):
    if not ( private_key_pem_path.exists() and root_crt_path.exists() ):
        # From version 1.8 the cert storage path is changed
        # if old key paris not exist create new ones
        # if old key pairs exist in old place copy it to new place
        if not (old_crt_path.exists() and old_private_key_pem_path.exists()):
            private_key_pem_path.parent.mkdir(parents=True, exist_ok=True)
            root_crt_path.parent.mkdir(parents=True, exist_ok=True)

            empty_subj = "/"
            create_root_cert(empty_subj, key_path=private_key_pem_path, cert_path=root_crt_path)
            mark_file(private_key_pem_path)
            mark_file(root_crt_path)
        else:
            shutil.move(old_crt_path, root_crt_path)
            shutil.move(old_private_key_pem_path, private_key_pem_path)

    if not check_permission(root_crt_path, uid=DEFAULT_UID, gid=DEFAULT_GID):
        os.chown(root_crt_path, DEFAULT_UID, DEFAULT_GID)

    if not check_permission(private_key_pem_path, uid=DEFAULT_UID, gid=DEFAULT_GID):
        os.chown(private_key_pem_path, DEFAULT_UID, DEFAULT_GID)


def prepare_trust_ca(config_dict):
    if shared_cert_dir.exists():
        shutil.rmtree(shared_cert_dir)
    shared_cert_dir.mkdir(parents=True, exist_ok=True)

    internal_ca_src = internal_tls_dir.joinpath(internal_ca_filename)
    ca_bundle_src = config_dict.get('registry_custom_ca_bundle_path')
    for src_path, dst_filename in (
        (internal_ca_src, internal_ca_filename),
        (ca_bundle_src, storage_ca_bundle_filename)):
        logging.info('copy {} to shared trust ca dir as name {} ...'.format(src_path, dst_filename))
        # check if source file valied
        if not src_path:
            continue
        real_src_path = get_realpath(str(src_path))
        if not real_src_path.exists():
            logging.info('ca file {} is not exist'.format(real_src_path))
            continue
        if not real_src_path.is_file():
            logging.info('{} is not file'.format(real_src_path))
            continue

        dst_path = shared_cert_dir.joinpath(dst_filename)

        # copy src to dst
        shutil.copy2(real_src_path, dst_path)

        # change ownership and permission
        mark_file(dst_path, mode=0o644)
