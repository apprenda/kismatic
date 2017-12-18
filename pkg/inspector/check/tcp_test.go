package check

import "testing"

func TestGetProcNameFromSockStatLine(t *testing.T) {
	tests := []struct {
		line         string
		expectedProc string
	}{
		{
			line:         `0      128               *:80                            *:*                   users:(("nginx",pid=30729,fd=10),("nginx",pid=30728,fd=10),("nginx",pid=30721,fd=10))`,
			expectedProc: "nginx",
		},
		{
			line:         `0      128              :::6443                         :::*                   users:(("kube-apiserver",pid=21199,fd=59))`,
			expectedProc: "kube-apiserver",
		},
		{
			line:         `0      128              :::2379                         :::*                   users:(("docker-proxy",pid=18718,fd=4))`,
			expectedProc: "docker-proxy",
		},
		{
			line:         `0      128              :::10251                        :::*                   users:(("kube-scheduler",pid=21506,fd=12))`,
			expectedProc: "kube-scheduler",
		},
		{
			line:         `0      128                                                              :::10251                                                                        :::*                   users:(("kube-scheduler",pid=21506,fd=12))`,
			expectedProc: "kube-scheduler",
		},
		{
			line:         `0      128                                                              :::10254                                                                        :::*                   users:(("nginx-ingress-c",pid=30704,fd=6))`,
			expectedProc: "nginx-ingress-c",
		},
	}
	for _, test := range tests {
		t.Run(test.expectedProc, func(t *testing.T) {
			got, err := getProcNameFromTCPSockStatLine(test.line)
			if err != nil {
				t.Errorf("error getting process: %v", err)
			}
			if got != test.expectedProc {
				t.Errorf("expected process name to be %q, but got %q", test.expectedProc, got)
			}
		})
	}
}
