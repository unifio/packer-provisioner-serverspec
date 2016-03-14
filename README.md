Packer Serverspec Provisioner
================================

[![Circle CI](https://circleci.com/gh/unifio/packer-provisioner-serverspec.svg?style=svg)](https://circleci.com/gh/unifio/packer-provisioner-serverspec)

Verifies VM configuration with Serverspec running remotely on the build server

Installation
------------
Install the binary (you'll need ```git``` and ```go```):

```
$ go get github.com/unifio/packer-provisioner-serverspec
```
Copy the plugin into packer.d directory:

```
$ mkdir $HOME/.packer.d/plugins
$ cp $GOPATH/bin/packer-provisioner-serverspec $HOME/.packer.d/plugins

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
    "type": "serverspec",
    "rake_file": "Rakefile",
    "rake_task": "serverspec:all",
    "rake_env_vars": "$BUNDLE_GEMFILE=Gemfile"
  }]
}
```

Configuration
-------------

All configuration properties are **required**, except where noted.

### rake_file

The relative path to the Rakefile to be utilized.

### rake_task

Rake task to be executed.

### user (optional)

SSH user to be used to connect to the SSH server on the host machine to forward commands to the target machine.

### local_port (optional)

A system-chosen port is used when `local_port` is missing or empty.

### ssh_host_key_file (optional)

(string) - The SSH key that will be used to run the SSH server on the host machine to forward commands to the target machine. Serverspec connects to this server and will validate the identity of the server using the system known_hosts. The default behavior is to generate and use a onetime key. Host key checking is disabled if the key is generated.

### ssh_authorized_key_file (optional)

(string) - The SSH public key of the Serverspec ssh_user. The default behavior is to generate and use a onetime key. If this key is generated, the corresponding private key is passed to Serverspec with the `TARGET_KEY` environment variable.

### sftp_command (optional)

(string) - The command to run on the machine being provisioned by Packer to handle the SFTP protocol that the plug-in will use to transfer files. The command should read and write on stdin and stdout, respectively. Defaults to `/usr/lib/sftp-server -e`.
