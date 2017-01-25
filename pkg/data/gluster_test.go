package data

import "testing"

var tests = []string{
	`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cliOutput>
<opRet>0</opRet>
<opErrno>0</opErrno>
<opErrstr/>
<volQuota>
<limit>
<path>/</path>
<hard_limit>1073741824</hard_limit>
<soft_limit_percent>80%</soft_limit_percent>
<soft_limit_value>858993459</soft_limit_value>
<used_space>0</used_space>
<avail_space>1073741824</avail_space>
<sl_exceeded>No</sl_exceeded>
<hl_exceeded>No</hl_exceeded>
</limit>
</volQuota>
</cliOutput>`,
	`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cliOutput>
<opRet>-1</opRet>
<opErrno>30800</opErrno>
<opErrstr>Volume storage01 does not exist</opErrstr>
<cliOp>volQuota</cliOp>
</cliOutput>`,
}

func TestUnmarshalVolumeQuota(t *testing.T) {
	for _, test := range tests {
		quota, err := UnmarshalVolumeQuota(test)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if quota == nil {
			t.Fatal("did not expect for quota to be nil")
		}
	}
}
