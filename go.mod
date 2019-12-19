module github.com/unifio/packer-provisioner-serverspec

require (
	github.com/hashicorp/hcl/v2 v2.0.0
	github.com/hashicorp/packer v1.5.0
	github.com/zclconf/go-cty v1.1.2-0.20191126233707-f0f7fd24c4af
	golang.org/x/crypto v0.0.0-20191202143827-86a70503ff7e
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

go 1.13
