import yaml
import click
import importlib
import os
from collections import deque

class MigratioNotFound(Exception): ...

class MigrationVersion:
    '''
    The version used to migration

    Arttribute:
        name(str): version name like `1.0.0`
        module: the python module object for a specific migration which contains migrate info, codes and templates
        down_versions(list): previous versions that can migrated to this version
    '''
    def __init__(self, version: str):
        self.name = version
        self.module = importlib.import_module("migrations.version_{}".format(version.replace(".","_")))

    @property
    def down_versions(self):
        return self.module.down_revisions

def read_conf(path):
    with open(path) as f:
        try:
            d = yaml.safe_load(f)
        except Exception as e:
            click.echo("parse config file err, make sure your harbor config version is above 1.8.0", e)
            exit(-1)
    return d

def search(input_version: str, target_version: str) -> list :
    """
    Find the migration path by BFS
    Args:
        input_version(str): The version migration start from
        target_version(str): The target version migrated to
    Returns:
        list: the module of migrations in the upgrade path
    """
    upgrade_path = []
    next_version, visited, q = {}, set(), deque()
    q.append(target_version)
    found = False
    while q: # BFS to find a valid path
        version = MigrationVersion(q.popleft())
        visited.add(version.name)
        if version.name == input_version:
            found = True
            break # break loop cause migration path found
        for v in version.down_versions:
            next_version[v] = version.name
            if v not in (visited.union(q)):
                q.append(v)

    if not found:
        raise MigratioNotFound('no migration path found to target version')

    current_version = MigrationVersion(input_version)
    while current_version.name != target_version:
        current_version = MigrationVersion(next_version[current_version.name])
        upgrade_path.append(current_version)
    return list(map(lambda x: x.module, upgrade_path))