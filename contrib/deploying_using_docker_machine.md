### Deploying Harbor using Docker Machine

Docker Machine allows you to deploy your containers to several cloud providers or on premises using a unified interface.
To deploy Harbor using Docker Machine, first create a virtual machine using Docker Machine.

This example will use DigitalOcean, but you can use VMware vCloud Air, AWS or Azure as well. Please see the list of supported drivers in the [Docker Machine driver documentation](https://docs.docker.com/machine/drivers/).

```
$ docker-machine create --driver digitalocean --digitalocean-access-token <youraccesstoken> harbor.mydomain.com
```

After the machine has been created successfully, you need to create a DNS entry at your provider for e.g. harbor.mydomain.com using the IP address for the machine we just created.
You can get this IP address using:

```
$ docker-machine ip harbor.mydomain.com
```

Copy `Deploy/harbor.yml.tmp` to `Deploy/harbor.yml` and make sure to change the `hostname` in `Deploy/harbor.yml` to `harbor.mydomain.com`, configure everything else according to the [Harbor Installation Guide](../docs/installation_guide.md) and run `prepare`.

Now, activate the created Docker Machine instance:

`$ eval $(docker-machine env harbor.mydomain.com)`

From within the `Deploy` directory, next copy the contents of the `config` directory to the machine.
First, get your local path to the `Deploy` directory:

```
$ echo $PWD
```

This will give you something like this:

```
/home/<yourusername>/src/harbor/Deploy
```

Then create this directory structure on the remote machine and copy the local files to the remote folders:

```
$ docker-machine ssh harbor.mydomain.com 'mkdir -p /home/<yourusername>/src/harbor/Deploy/config
$ docker-machine scp -r ./config harbor.mydomain.com:$PWD
```

Next, build your Harbor images:

```
$ docker-compose build
```

And finally, spin up your Harbor containers:

```
$ docker-compose up -d
```

Now you should be able to browse `http://harbor.mydomain.com`.
