import os
import shutil
from g import config_dir, templates_dir, data_dir, DEFAULT_GID, DEFAULT_UID
from utils.jinja import render_jinja
from utils.misc import prepare_dir, generate_random_string

core_config_dir = os.path.join(config_dir, "core", "certificates")
core_env_template_path = os.path.join(templates_dir, "core", "env.jinja")
core_conf_env = os.path.join(config_dir, "core", "env")
core_conf_template_path = os.path.join(templates_dir, "core", "app.conf.jinja")
core_conf = os.path.join(config_dir, "core", "app.conf")

ca_download_dir = os.path.join(data_dir, 'ca_download')


def prepare_core(config_dict, with_notary, with_clair, with_trivy, with_chartmuseum):
    prepare_dir(ca_download_dir, uid=DEFAULT_UID, gid=DEFAULT_GID)
    prepare_dir(core_config_dir)
    # Render Core
    # set cache for chart repo server
    # default set 'memory' mode, if redis is configured then set to 'redis'
    if len(config_dict['redis_url_core']) > 0:
        chart_cache_driver = "redis"
    else:
        chart_cache_driver = "memory"

    render_jinja(
        core_env_template_path,
        core_conf_env,
        chart_cache_driver=chart_cache_driver,
        with_notary=with_notary,
        with_clair=with_clair,
        with_trivy=with_trivy,
        with_chartmuseum=with_chartmuseum,
        csrf_key=generate_random_string(32),
        **config_dict)

    render_jinja(
        core_conf_template_path,
        core_conf,
        uid=DEFAULT_UID,
        gid=DEFAULT_GID)


def copy_core_config(core_templates_path, core_config_path):
    shutil.copyfile(core_templates_path, core_config_path)
    print("Generated configuration file: %s" % core_config_path)
