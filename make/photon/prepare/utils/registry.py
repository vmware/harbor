import copy
import os
import subprocess
from g import config_dir, templates_dir, DEFAULT_GID, DEFAULT_UID, data_dir
from urllib.parse import urlsplit
from utils.jinja import render_jinja
from utils.misc import prepare_dir

registry_config_dir = os.path.join(config_dir, "registry")
registry_config_template_path = os.path.join(templates_dir, "registry", "config.yml.jinja")
registry_conf = os.path.join(config_dir, "registry", "config.yml")
registry_passwd_path = os.path.join(config_dir, "registry", "passwd")
registry_data_dir = os.path.join(data_dir, 'registry')

levels_map = {
    'debug': 'debug',
    'info': 'info',
    'warning': 'warn',
    'error': 'error',
    'fatal': 'fatal'
}


def prepare_registry(config_dict):
    prepare_dir(registry_data_dir, uid=DEFAULT_UID, gid=DEFAULT_GID)
    prepare_dir(registry_config_dir)

    gen_passwd_file(config_dict)
    storage_provider_info = get_storage_provider_info(
        config_dict['storage_provider_name'],
        config_dict['storage_provider_config'])

    # process redis info
    redis_ops = parse_redis(config_dict['redis_url_reg'])

    render_jinja(
        registry_config_template_path,
        registry_conf,
        uid=DEFAULT_UID,
        gid=DEFAULT_GID,
        level=levels_map[config_dict['log_level']],
        storage_provider_info=storage_provider_info,
        **config_dict, **redis_ops)


def parse_redis(redis_url):
    u = urlsplit(redis_url)
    if not u.scheme or u.scheme == 'redis':
        return {
            'redis_host': u.netloc.split('@')[-1],
            'redis_password': u.password or '',
            'redis_db_index_reg': u.path and int(u.path[1:]) or 0,
        }
    elif u.scheme == 'redis+sentinel':
        return {
            'sentinel_master_set': u.path.split('/')[1],
            'redis_host': u.netloc.split('@')[-1],
            'redis_password': u.password or '',
            'redis_db_index_reg': len(u.path.split('/')) == 3 and int(u.path.split('/')[2]) or 0,
        }
    else:
        raise Exception('bad redis url for registry:' + redis_url)

def get_storage_provider_info(provider_name, provider_config):
    provider_config_copy = copy.deepcopy(provider_config)
    if provider_name == "filesystem":
        if not (provider_config_copy and ('rootdirectory' in provider_config_copy)):
            provider_config_copy['rootdirectory'] = '/storage'
    if provider_name == 'gcs' and provider_config_copy.get('keyfile'):
        provider_config_copy['keyfile'] = '/etc/registry/gcs.key'
    # generate storage configuration section in yaml format
    storage_provider_conf_list = [provider_name + ':']
    for config in provider_config_copy.items():
        if config[1] is None:
            value = ''
        elif config[1] == True:
            value = 'true'
        else:
            value = config[1]
        storage_provider_conf_list.append('{}: {}'.format(config[0], value))
    storage_provider_info = ('\n' + ' ' * 4).join(storage_provider_conf_list)
    return storage_provider_info


def gen_passwd_file(config_dict):
    return subprocess.call(["/usr/bin/htpasswd", "-bcB", registry_passwd_path, config_dict['registry_username'],
                            config_dict['registry_password']], stdout=subprocess.DEVNULL, stderr=subprocess.STDOUT)
