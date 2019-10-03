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

The connection to the builder is facilitated via a local SSH proxy. The integration requires that your `spec_helper` minimally exposes the following parameters:

```ruby
require 'serverspec'
require 'net/ssh'

options = Net::SSH::Config.for(host, [])
options[:user] = ENV['TARGET_USER']
options[:keys] = ENV['TARGET_KEY']
options[:host_name] = ENV['TARGET_HOST']
options[:port] = ENV['TARGET_PORT']
options[:verify_host_key] = :never unless ENV['SERVERSPEC_HOST_KEY_CHECKING'] =~ (/^(true|t|yes|y|1)$/i)

set :host,         options[:host_name]
set :ssh_options,  options
set :backend,      :ssh
set :display_sudo, true
set :request_pty,  true
```

Configuration
-------------

All configuration properties are **required**, except where noted.

### rake_file

The relative path to the Rakefile to be utilized.

### rake_task

Rake task to be executed.

### command (optional)

Command to be used to execute Serverspec. Defaults to `rake`.

** Note: `bundle exec` will not work if passed into the `command` parameter. If you want to achieve the same context isolation, add `require 'bundler/setup'` to your Rakefile and pass `$BUNDLE_GEMFILE` to the `rake_env_vars` parameter as illustrated in the example above.

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
