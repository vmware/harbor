# pylint: disable=no-value-for-parameter

import sys
import logging

import click

from utils.misc import delfile
from utils.configs import validate, parse_yaml_config
from utils.cert import prepare_registry_ca, SSL_CERT_KEY_PATH, SSL_CERT_PATH, get_secret_key, prepare_trust_ca
from utils.db import prepare_db
from utils.jobservice import prepare_job_service
from utils.registry import prepare_registry
from utils.registry_ctl import prepare_registry_ctl
from utils.core import prepare_core
from utils.notary import prepare_notary
from utils.log import prepare_log_configs
from utils.clair import prepare_clair
from utils.clair_adapter import prepare_clair_adapter
from utils.chart import prepare_chartmuseum
from utils.docker_compose import prepare_docker_compose
from utils.nginx import prepare_nginx, nginx_confd_dir
from utils.redis import prepare_redis
from utils.internal_tls import prepare_tls
from utils.trivy_adapter import prepare_trivy_adapter
from g import (config_dir, input_config_path, private_key_pem_path, root_crt_path, secret_key_dir,
old_private_key_pem_path, old_crt_path)

@click.command()
@click.option('--conf', default=input_config_path, help="the path of Harbor configuration file")
@click.option('--with-notary', is_flag=True, help="the Harbor instance is to be deployed with notary")
@click.option('--with-clair', is_flag=True, help="the Harbor instance is to be deployed with clair")
@click.option('--with-trivy', is_flag=True, help="the Harbor instance is to be deployed with Trivy")
@click.option('--with-chartmuseum', is_flag=True, help="the Harbor instance is to be deployed with chart repository supporting")
def prepare(conf, with_notary, with_clair, with_trivy, with_chartmuseum):

    delfile(config_dir)
    config_dict = parse_yaml_config(conf, with_notary=with_notary, with_clair=with_clair, with_trivy=with_trivy, with_chartmuseum=with_chartmuseum)
    try:
        validate(config_dict, notary_mode=with_notary)
    except Exception as e:
        click.echo('Error happened in config validation...')
        logging.error(e)
        sys.exit(-1)

    prepare_log_configs(config_dict)
    prepare_nginx(config_dict)
    prepare_core(config_dict, with_notary=with_notary, with_clair=with_clair, with_trivy=with_trivy, with_chartmuseum=with_chartmuseum)
    prepare_registry(config_dict)
    prepare_registry_ctl(config_dict)
    prepare_db(config_dict)
    prepare_job_service(config_dict)
    prepare_redis(config_dict)
    prepare_tls(config_dict)
    prepare_trust_ca(config_dict)

    get_secret_key(secret_key_dir)

    #  If Customized cert enabled
    prepare_registry_ca(
        private_key_pem_path=private_key_pem_path,
        root_crt_path=root_crt_path,
        old_private_key_pem_path=old_private_key_pem_path,
        old_crt_path=old_crt_path)

    if with_notary:
        prepare_notary(config_dict, nginx_confd_dir, SSL_CERT_PATH, SSL_CERT_KEY_PATH)

    if with_clair:
        prepare_clair(config_dict)
        prepare_clair_adapter(config_dict)

    if with_trivy:
        prepare_trivy_adapter(config_dict)

    if with_chartmuseum:
        prepare_chartmuseum(config_dict)

    prepare_docker_compose(config_dict, with_clair, with_trivy, with_notary, with_chartmuseum)
