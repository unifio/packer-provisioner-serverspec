module github.com/unifio/packer-provisioner-serverspec

require (
	github.com/hashicorp/hcl/v2 v2.16.0
	github.com/hashicorp/packer v1.6.0
	github.com/zclconf/go-cty v1.12.1
	golang.org/x/crypto v0.0.0-20220517005047-85d78b3ac167
)

require (
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/apparentlymart/go-textseg/v13 v13.0.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	golang.org/x/text v0.7.0 // indirect
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

go 1.18
