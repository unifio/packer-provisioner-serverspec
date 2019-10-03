package main

import (
	"github.com/hashicorp/packer/packer/plugin"
	"github.com/unifio/packer-provisioner-serverspec/serverspec"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterProvisioner(new(serverspec.Provisioner))
	server.Serve()
}
