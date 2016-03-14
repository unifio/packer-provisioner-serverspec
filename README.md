Packer Fabric Provisioner
================================

[![Circle CI](https://circleci.com/gh/unifio/packer-provisioner-fabric.svg?style=svg)](https://circleci.com/gh/unifio/packer-provisioner-fabric)

Provisions VM with Fabric running remotely on the build server

Installation
------------
Install the binary (you'll need ```git``` and ```go```):

```
$ go get github.com/unifio/packer-provisioner-fabric
```
Copy the plugin into packer.d directory:

```
$ mkdir $HOME/.packer.d/plugins
$ cp $GOPATH/bin/packer-provisioner-fabric $HOME/.packer.d/plugins

```
Usage
-----

Add the provisioner to your packer template:

```json
{
  "variables": {
    "version":  "0.0.1",
    "box_name": "my-cool-project"
  },
  "builders": [ ... ],
  "provisioners": [{
    "type": "fabric",
    "fab_file": "fabfile.py",
    "fab_tasks": "test"
  }]
}
```

Configuration
-------------

All configuration properties are **required**, except where noted.

### fab_file

The relative path to the fabfile to be utilized.

### fab_tasks

Comma separated lit of tasks to be executed.

### user (optional)

SSH user to be used to connect to the SSH server on the host machine to forward commands to the target machine.

### local_port (optional)

A system-chosen port is used when `local_port` is missing or empty.

### ssh_host_key_file (optional)

(string) - The SSH key that will be used to run the SSH server on the host machine to forward commands to the target machine. Fabric connects to this server and will validate the identity of the server using the system known_hosts. The default behavior is to generate and use a onetime key. Host key checking is disabled if the key is generated.

### ssh_authorized_key_file (optional)

(string) - The SSH public key of the Fabric ssh_user. The default behavior is to generate and use a onetime key. If this key is generated, the corresponding private key is passed to `fab` with the `-i` option.
