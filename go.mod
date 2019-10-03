module github.com/unifio/packer-provisioner-serverspec

require (
	github.com/hashicorp/packer v1.4.4
	golang.org/x/crypto v0.0.0-20190605123033-f99c8df09eb5
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

go 1.13
