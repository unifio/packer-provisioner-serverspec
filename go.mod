module github.com/unifio/packer-provisioner-serverspec

require (
	github.com/hashicorp/hcl/v2 v2.3.0
	github.com/hashicorp/packer v1.5.4
	github.com/zclconf/go-cty v1.2.1
	golang.org/x/crypto v0.0.0-20200117160349-530e935923ad
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

go 1.13
