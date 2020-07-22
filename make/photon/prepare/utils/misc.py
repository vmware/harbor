import os
import secrets
import string
import sys
from pathlib import Path
from functools import wraps
from g import DEFAULT_UID, DEFAULT_GID, host_root_dir


# To meet security requirement
# By default it will change file mode to 0600, and make the owner of the file to 10000:10000
def mark_file(path, mode=0o600, uid=DEFAULT_UID, gid=DEFAULT_GID):
    if mode > 0:
        os.chmod(path, mode)
    if uid > 0 and gid > 0:
        os.chown(path, uid, gid)


def validate(conf, **kwargs):
    # Protocol validate
    protocol = conf.get("configuration", "ui_url_protocol")
    if protocol != "https" and kwargs.get('notary_mode'):
        raise Exception(
            "Error: the protocol must be https when Harbor is deployed with Notary")
    if protocol == "https":
        if not conf.has_option("configuration", "ssl_cert"):
            raise Exception(
                "Error: The protocol is https but attribute ssl_cert is not set")
        cert_path = conf.get("configuration", "ssl_cert")
        if not os.path.isfile(cert_path):
            raise Exception(
                "Error: The path for certificate: %s is invalid" % cert_path)
        if not conf.has_option("configuration", "ssl_cert_key"):
            raise Exception(
                "Error: The protocol is https but attribute ssl_cert_key is not set")
        cert_key_path = conf.get("configuration", "ssl_cert_key")
        if not os.path.isfile(cert_key_path):
            raise Exception(
                "Error: The path for certificate key: %s is invalid" % cert_key_path)

    # Storage validate
    valid_storage_drivers = ["filesystem",
                             "azure", "gcs", "s3", "swift", "oss"]
    storage_provider_name = conf.get(
        "configuration", "registry_storage_provider_name").strip()
    if storage_provider_name not in valid_storage_drivers:
        raise Exception("Error: storage driver %s is not supported, only the following ones are supported: %s" % (
            storage_provider_name, ",".join(valid_storage_drivers)))

    storage_provider_config = conf.get(
        "configuration", "registry_storage_provider_config").strip()
    if storage_provider_name != "filesystem":
        if storage_provider_config == "":
            raise Exception(
                "Error: no provider configurations are provided for provider %s" % storage_provider_name)

def validate_crt_subj(dirty_subj):
    subj_list = [item for item in dirty_subj.strip().split("/") \
        if len(item.split("=")) == 2 and len(item.split("=")[1]) > 0]
    return "/" + "/".join(subj_list)


def generate_random_string(length):
    return ''.join(secrets.choice(string.ascii_letters + string.digits) for _ in range(length))


def prepare_dir(root: str, *args, **kwargs) -> str:
    gid, uid = kwargs.get('gid'), kwargs.get('uid')
    absolute_path = Path(os.path.join(root, *args))
    if absolute_path.is_file():
        raise Exception('Path exists and the type is regular file')
    mode = kwargs.get('mode') or 0o755

    # we need make sure this dir has the right permission
    if not absolute_path.exists():
        absolute_path.mkdir(mode=mode, parents=True)
    elif not check_permission(absolute_path, mode=mode):
         absolute_path.chmod(mode)

    # if uid or gid not None, then change the ownership of this dir
    if not(gid is None and uid is None):
        dir_uid, dir_gid = absolute_path.stat().st_uid, absolute_path.stat().st_gid
        if uid is None:
            uid = dir_uid
        if gid is None:
            gid = dir_gid
        # We decide to recursively chown only if the dir is not owned by correct user
        # to save time if the dir is extremely large
        if not check_permission(absolute_path, uid, gid):
            recursive_chown(absolute_path, uid, gid)

    return str(absolute_path)



def delfile(src):
    if os.path.isfile(src):
        try:
            os.remove(src)
            print("Clearing the configuration file: %s" % src)
        except Exception as e:
            print(e)
    elif os.path.isdir(src):
        for dir_name in os.listdir(src):
            dir_path = os.path.join(src, dir_name)
            delfile(dir_path)


def recursive_chown(path, uid, gid):
    os.chown(path, uid, gid)
    for root, dirs, files in os.walk(path):
        for d in dirs:
            os.chown(os.path.join(root, d), uid, gid)
        for f in files:
            os.chown(os.path.join(root, f), uid, gid)


def check_permission(path: str, uid:int = None, gid:int = None, mode:int = None):
    if not isinstance(path, Path):
        path = Path(path)
    if uid is not None and uid != path.stat().st_uid:
        return False
    if gid is not None and gid != path.stat().st_gid:
        return False
    if mode is not None and (path.stat().st_mode - mode) % 0o1000 != 0:
        return False
    return True


def owner_can_read(st_mode: int) -> bool:
    """
    Check if owner have the read permission of this st_mode
    """
    return True if st_mode & 0o400 else False


def other_can_read(st_mode: int) -> bool:
    """
    Check if other user have the read permission of this st_mode
    """
    return True if st_mode & 0o004 else False


# decorator actions
def stat_decorator(func):
    @wraps(func)
    def check_wrapper(*args, **kw):
        stat = func(*args, **kw)
        if stat == 0:
            print("Successfully called func: %s" % func.__name__)
        else:
            print("Failed to call func: %s" % func.__name__)
            sys.exit(1)
    return check_wrapper


def get_realpath(path: str) -> Path:
    """
    Return the real path in your host if you mounted your host's filesystem to /hostfs,
    or return the original path
    """

    if os.path.isdir(host_root_dir):
        return host_root_dir.joinpath(path.lstrip('/'))
    return Path(path)
