import pytest
import importlib

from utils.migration import search, MigratioNotFound

class mockModule:
    def __init__(self, revision: str, down_revisions: list):
        self.revision = revision
        self.down_revisions = down_revisions

def mock_import_module_loop(module_path: str):
    modules = {
    'migrations.version_1_9_0': mockModule('1.9.0', []),
    'migrations.version_1_10_0': mockModule('1.10.0', ['2.0.0']),
    'migrations.version_2_0_0': mockModule('2.0.0', ['1.10.0'])
    }
    return modules[module_path]

def mock_import_module_mission(module_path: str):
    modules = {
    'migrations.version_1_9_0': mockModule('1.9.0', []),
    'migrations.version_1_10_0': mockModule('1.10.0', []),
    'migrations.version_2_0_0': mockModule('2.0.0', ['1.10.0'])
    }
    return modules[module_path]

def mock_import_module_success(module_path: str):
    modules = {
    'migrations.version_1_9_0': mockModule('1.9.0', []),
    'migrations.version_1_10_0': mockModule('1.10.0', ['1.9.0']),
    'migrations.version_2_0_0': mockModule('2.0.0', ['1.10.0'])
    }
    return modules[module_path]

def mock_import_module_success_multi_downversion(module_path: str):
    modules = {
    'migrations.version_1_9_0': mockModule('1.9.0', []),
    'migrations.version_1_10_0': mockModule('1.10.0', ['1.9.0']),
    'migrations.version_1_10_1': mockModule('1.10.1', ['1.9.0']),
    'migrations.version_1_10_2': mockModule('1.10.2', ['1.9.0']),
    'migrations.version_2_0_0': mockModule('2.0.0', ['1.10.0', '1.10.1', '1.10.2'])
    }
    return modules[module_path]

@pytest.fixture
def mock_import_module_with_loop(monkeypatch):
    monkeypatch.setattr(importlib, "import_module", mock_import_module_loop)

@pytest.fixture
def mock_import_module_with_mission(monkeypatch):
    monkeypatch.setattr(importlib, "import_module", mock_import_module_mission)

@pytest.fixture
def mock_import_module_with_success(monkeypatch):
    monkeypatch.setattr(importlib, "import_module", mock_import_module_success)

@pytest.fixture
def mock_import_module_with_success_multi_downversion(monkeypatch):
    monkeypatch.setattr(importlib, "import_module", mock_import_module_success_multi_downversion)

def test_search_loop(mock_import_module_with_loop):
    with pytest.raises(Exception):
        search('1.9.0', '2.0.0')

def test_search_mission(mock_import_module_with_mission):
    with pytest.raises(MigratioNotFound):
        search('1.9.0', '2.0.0')

def test_search_success(mock_import_module_with_success):
    migration_path = search('1.9.0', '2.0.0')
    assert migration_path[0].revision == '1.10.0'
    assert migration_path[1].revision == '2.0.0'

def test_search_success_multi_downversion(mock_import_module_with_success_multi_downversion):
    migration_path = search('1.9.0', '2.0.0')
    print(migration_path)
    assert migration_path[0].revision == '1.10.2'
    assert migration_path[1].revision == '2.0.0'
