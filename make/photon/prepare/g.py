import os
from pathlib import Path

## Const
DEFAULT_UID = 10000
DEFAULT_GID = 10000

PG_UID = 999
PG_GID = 999

REDIS_UID = 999
REDIS_GID = 999

## Global variable
templates_dir = "/usr/src/app/templates"

host_root_dir = Path('/hostfs')

base_dir = '/harbor_make'
config_dir = Path('/config')
data_dir = Path('/data')

secret_dir = data_dir.joinpath('secret')
secret_key_dir = secret_dir.joinpath('keys')
trust_ca_dir = secret_dir.joinpath('keys', 'trust_ca')
internal_tls_dir = secret_dir.joinpath('tls')

storage_ca_bundle_filename = 'storage_ca_bundle.crt'
internal_ca_filename = 'harbor_internal_ca.crt'

old_private_key_pem_path = Path('/config/core/private_key.pem')
old_crt_path = Path('/config/registry/root.crt')

private_key_pem_path = secret_dir.joinpath('core', 'private_key.pem')
root_crt_path = secret_dir.joinpath('registry', 'root.crt')

config_file_path = '/compose_location/harbor.yml'
input_config_path = '/input/harbor.yml'
versions_file_path = Path('/usr/src/app/versions')

cert_dir = config_dir.joinpath("nginx", "cert")
core_cert_dir = config_dir.joinpath("core", "certificates")
shared_cert_dir = config_dir.joinpath("shared", "trust-certificates")

INTERNAL_NO_PROXY_DN = {
    '127.0.0.1',
    'localhost',
    '.local',
    '.internal',
    'log',
    'db',
    'redis',
    'nginx',
    'core',
    'portal',
    'postgresql',
    'jobservice',
    'registry',
    'registryctl',
    'clair',
    'chartmuseum',
    'notary-server',
    'notary-signer',
    'clair-adapter',
    'trivy-adapter',
    }
