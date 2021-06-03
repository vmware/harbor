import os

from g import templates_dir
from .configs import parse_versions
from .jinja import render_jinja

docker_compose_template_path = os.path.join(templates_dir, 'docker_compose', 'docker-compose.yml.jinja')
docker_compose_yml_path = '/compose_location/docker-compose.yml'

# render docker-compose
def prepare_docker_compose(configs, with_trivy, with_notary, with_chartmuseum):
    versions = parse_versions()
    VERSION_TAG = versions.get('VERSION_TAG') or 'dev'

    rendering_variables = {
        'version': VERSION_TAG,
        'reg_version': VERSION_TAG,
        'redis_version': VERSION_TAG,
        'notary_version': VERSION_TAG,
        'trivy_adapter_version': VERSION_TAG,
        'chartmuseum_version': VERSION_TAG,
        'data_volume': configs['data_volume'],
        'log_location': configs['log_location'],
        'protocol': configs['protocol'],
        'http_port': configs['http_port'],
        'external_redis': configs['external_redis'],
        'external_database': configs['external_database'],
        'with_notary': with_notary,
        'with_trivy': with_trivy,
        'with_chartmuseum': with_chartmuseum
    }

    # if configs.get('registry_custom_ca_bundle_path'):
    #     rendering_variables['registry_custom_ca_bundle_path'] = configs.get('registry_custom_ca_bundle_path')
    #     rendering_variables['custom_ca_required'] = True

    # for gcs
    storage_config = configs.get('storage_provider_config') or {}
    if storage_config.get('keyfile') and configs['storage_provider_name'] == 'gcs':
        rendering_variables['gcs_keyfile'] = storage_config['keyfile']

    # for http
    if configs['protocol'] == 'https':
        rendering_variables['cert_key_path'] = configs['cert_key_path']
        rendering_variables['cert_path'] = configs['cert_path']
        rendering_variables['https_port'] = configs['https_port']

    # internal cert pairs
    rendering_variables['internal_tls'] = configs['internal_tls']

    # for uaa
    uaa_config = configs.get('uaa') or {}
    if uaa_config.get('ca_file'):
        rendering_variables['uaa_ca_file'] = uaa_config['ca_file']

    # for log
    log_ep_host = configs.get('log_ep_host')
    if log_ep_host:
        rendering_variables['external_log_endpoint'] = True

    # for metrics
    metric = configs.get('metric')
    if metric:
        rendering_variables['metric'] = metric

    # for ipv6
    ipv6 = configs.get('ipv6')
    if ipv6:
        rendering_variables['ipv6'] = ipv6

    render_jinja(docker_compose_template_path, docker_compose_yml_path,  mode=0o644, **rendering_variables)
