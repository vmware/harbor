---
title: Upgrade Harbor and Migrate Data
weight: 45
---

This guide covers upgrade and migration to version 2.0.0. This guide only covers migration from v1.9.x and later to the current version. If you are upgrading from an earlier version, refer to the migration guide in the `release-1.9.0` branch to upgrade to v1.9.x first, then follow this guide to perform the migration to this version.

If you are upgrading a Harbor instance that you deployed with Helm, see [Upgrading Harbor Deployed with Helm](helm-upgrade.md).

When upgrading an existing Harbor instance to a newer version, you might need to migrate the settings in `harbor.yml`.
Since the migration might alter the database schema and the settings of `harbor.yml`, you should **always** back up your data before any migration.

## Notes

- Again, you must back up your data before any data migration.
- In version 1.9.0, some containers are started by `non-root`. This does not pose problems if you are upgrading an officially released version of Harbor, but if you have deployed a customized instance of Harbor, you might encounter permission issues.
- In previous releases, user roles took precedence over group roles in a project. In this version, user roles and group roles are combined so that the user has whichever set of permissions is highest. This might cause the roles of certain users to change during upgrade.
- With the introduction of storage and artifact quotas in version 1.9.0, migration from 1.8.x might take a few minutes. This is because the `core` walks through all blobs in the registry and populates the database with information about the layers and artifacts in projects.
- With the introduction of storage and artifact quotas in version 1.9.0, replication between version 1.9.0 and a previous version of Harbor does not work. You must upgrade all Harbor nodes to 1.9.0 if you have configured replication between them.

## Upgrading Harbor and Migrating Data

1. Log in to the Harbor host and, if it is still running, stop and remove the existing Harbor instance.

    ```sh
    cd harbor
    docker-compose down
    ```

1. Back up Harbor's current files so that you can roll back to the current version if necessary.

    ```sh
    mv harbor /my_backup_dir/harbor
    ```

1. Back up the database, which by default is in the directory `/data/database`.

    ```sh
    cp -r /data/database /my_backup_dir/
    ```

1. Get the latest Harbor release package from [https://github.com/goharbor/harbor/releases](https://github.com/goharbor/harbor/releases).
1. Before upgrading Harbor, perform migration.

    The migration tool is in harbor-prepare tools delivered as a docker image. You can pull the image from docker hub. in the following command:

    ```sh
    docker pull goharbor/prepare:[tag]
    ```

    Alternatively, if you are using an offline installer package, you can load it from the image tarball that is included in the offline installer package. Replace [tag] with the new Harbor version, for example v1.10.0, in the following command:

    ```sh
    tar zxf <offline package>
    docker image load -i harbor/harbor.[version].tar.gz
    ```

1. Copy the `harbor.yml.tmp` to `harbor.yml` and upgrade it.

    ```sh
    docker run -it --rm -v /:/hostfs goharbor/prepare:[tag] migrate -i ${path to harbor.yml}
    ```

    **NOTE:** The schema upgrade and data migration of the database is performed by core when Harbor starts. If the migration fails, check the core log to debug.

3. In the `./harbor` directory, run the `./install.sh` script to install the new Harbor instance.

   To install Harbor with components such as Notary, Clair, and chartmuseum, see [Run the Installer Script](../../install-config/run-installer-script.md) for more information.
   
If you need to roll back to the previous version of Harbor, see [Roll Back from an Upgrade](roll-back-upgrade.md).
