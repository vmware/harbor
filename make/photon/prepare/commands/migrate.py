import os, sys, importlib, shutil, glob
from packaging import version

import click

from utils.misc import get_realpath
from utils.migration import read_conf, search
from migrations import accept_versions

@click.command()
@click.option('-i', '--input', 'input_', required=True, help="The path of original config file")
@click.option('-o', '--output', default='', help="the path of output config file")
@click.option('-t', '--target', default='2.0.0', help="target version of input path")
def migrate(input_, output, target):
    """
    migrate command will migrate config file style to specific version
    :input_: is the path of the original config file
    :output: is the destination path of config file, the generated configs will storage in it
    :target: is the the target version of config file will upgrade to
    """
    if target not in accept_versions:
        click.echo('target version {} not supported'.format(target))
        sys.exit(-1)

    if not output:
        output = input_
    input_path = get_realpath(input_)
    output_path = get_realpath(output)

    configs = read_conf(input_path)
    input_version = configs.get('_version')
    if version.parse(input_version) < version.parse('1.9.0'):
        click.echo('the version {} not supported, make sure the version in input file above 1.8.0'.format(input_version))
        sys.exit(-1)
    if input_version == target:
        click.echo("Version of input harbor.yml is identical to target {}, no need to upgrade".format(input_version))
        sys.exit(0)

    current_input_path = input_path
    for m in search(input_version, target):
        current_output_path = "harbor.yml.{}.tmp".format(m.revision)
        click.echo("migrating to version {}".format(m.revision))
        m.migrate(current_input_path, current_output_path)
        current_input_path = current_output_path
    shutil.copy(current_input_path, output_path)
    click.echo("Written new values to {}".format(output))
    for tmp_f in glob.glob("harbor.yml.*.tmp"):
        os.remove(tmp_f)

