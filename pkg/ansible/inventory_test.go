package ansible

import "testing"

func TestInventoryINIGeneration(t *testing.T) {
	inv := Inventory{
		Roles: []Role{
			{
				Name: "etcd",
				Nodes: []Node{
					{
						Host:          "etcd01",
						PublicIP:      "10.0.0.1",
						InternalIP:    "192.168.0.11",
						SSHPrivateKey: "id_rsa",
						SSHPort:       2222,
						SSHUser:       "alice",
					},
				},
			},
			{
				Name: "master",
				Nodes: []Node{
					{
						Host:          "master01",
						PublicIP:      "10.0.0.2",
						InternalIP:    "192.168.0.12",
						SSHPrivateKey: "id_rsa",
						SSHPort:       2222,
						SSHUser:       "alice",
					},
				},
			}, {
				Name: "worker",
				Nodes: []Node{
					{
						Host:          "worker01",
						PublicIP:      "10.0.0.3",
						InternalIP:    "192.168.0.13",
						SSHPrivateKey: "id_rsa",
						SSHPort:       2222,
						SSHUser:       "alice",
					},
					{
						Host:          "worker02",
						PublicIP:      "10.0.0.4",
						InternalIP:    "192.168.0.14",
						SSHPrivateKey: "id_rsa",
						SSHPort:       2222,
						SSHUser:       "alice and bob",
					},
				},
			},
		},
	}

	ini := string(inv.ToINI())

	expected := `[etcd]
"etcd01" ansible_host="10.0.0.1" internal_ipv4="192.168.0.11" ansible_ssh_private_key_file="id_rsa" ansible_port=2222 ansible_user="alice"
[master]
"master01" ansible_host="10.0.0.2" internal_ipv4="192.168.0.12" ansible_ssh_private_key_file="id_rsa" ansible_port=2222 ansible_user="alice"
[worker]
"worker01" ansible_host="10.0.0.3" internal_ipv4="192.168.0.13" ansible_ssh_private_key_file="id_rsa" ansible_port=2222 ansible_user="alice"
"worker02" ansible_host="10.0.0.4" internal_ipv4="192.168.0.14" ansible_ssh_private_key_file="id_rsa" ansible_port=2222 ansible_user="alice and bob"
`

	if ini != expected {
		t.Errorf("expected format differs from obtained format. Expected: \n%s\nGot: \n%s\n", expected, ini)
	}

}
