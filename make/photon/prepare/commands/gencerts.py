import os
import sys
import click
import pathlib
import logging
from subprocess import Popen, PIPE, STDOUT, CalledProcessError

from utils.cert import openssl_installed
from utils.misc import get_realpath

gen_tls_script = pathlib.Path(__file__).parent.parent.joinpath('scripts/gencert.sh').absolute()

@click.command()
@click.option('-p', '--path', required=True, type=str,help='the path to store generated cert files')
@click.option('-d', '--days', default='365', type=str, help='the expired time for cert')
def gencert(path, days):
    """
    gencert command will generate cert files for internal TLS
    """
    path = get_realpath(path)
    click.echo('Check openssl ...')
    if not openssl_installed():
        raise(Exception('openssl not installed'))

    click.echo("start generate internal tls certs")
    if not os.path.exists(path):
        click.echo('path {} not exist, create it...'.format(path))
        os.makedirs(path, exist_ok=True)
    with Popen([gen_tls_script, days], stdout=PIPE, stderr=STDOUT, cwd=path) as p:
        for line in p.stdout:
            click.echo(line, nl=False)
    if p.returncode != 0:
        raise CalledProcessError(p.returncode, p.args)
