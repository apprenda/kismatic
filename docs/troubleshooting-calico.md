# Troubleshooting Calico Networking

Calico relies on an agent running on each node for the purposes of initializing the Calico network on the node, updating routing information and applying policy (when enabled).

The Calico agent is packaged up as a container that starts multiple binaries that work together:
* Felix: Programs routes and network interfaces on the node for connectivity to and from workloads
* BIRD: BGP client that is used to sync routing information between nodes. BIRD picks up routes programmed by Felix, and distributes them via BGP.
* confd: Monitors etcd datastore for configuration changes

## Logs
The first thing to do when troubleshooting is to check the logs on the node that is experiencing issues. Looking at logs can help in ensuring that the system has been configured properly, and that the components themselves are OK.

Logs for the Calico components are located in:
* Felix: `/var/log/calico/felix.log`
* BIRD: `/var/log/calico/bird/current`
* confd: `/var/log/calico/confd/current`

The kubelet, which interfaces with Calico through CNI, can also emit logs with information about failures.

## Overlay mode

In Overlay mode, Calico uses IP-in-IP encapsulation, or tunneling, to flow packets between workloads. At a basic level, what this means is that the IP packets created by the workload are encapsulated by the kernel with another packet that has a different destination IP.

##### Scenario: I am able to access the workload from the node where it's running on, but not from other nodes.

1. Isolate two nodes that will be used for troubleshooting. The first node is a "server" node hosting a workload (assuming nginx), and a second node is a "client" node that is unable to communicate with the workload.
2. Use `ping` to verify connectivity between the client and server nodes.
3. Obtain the IP address of the tunnel interface on the server node: `ip addr show dev tunl0` (Referred to as `$SERVER_TUN_IP` hereinafter).
4. On the client node, verify that you are able to ping the server tunnel interface: `ping $SERVER_TUN_IP`
5. If unable to ping, verify the route is set up properly on the client node using `ip route` (You may choose to run `ip route | grep $SERVER_TUN_IP`).
For example, assuming `$SERVER_TUN_IP=172.16.252.64` and `$SERVER_IP=10.99.224.143`, a correct routing table entry would be:
`172.16.252.64/26 via 10.99.224.143 dev tunl0  proto bird onlink`.
Verify that the rule lists `tunl0` as the interface.

If you are unable to ping the `tunl0` interface on another node, this might indicate an issue with the network.
In order to troubleshoot the network, we will use `tcpdump` and `ping` to try and find where the packets are being
dropped or getting lost.

#### Verify client->server connectivity

On the client node, start capturing packets destined to the server node in the background: `tcpdump -n -vvv -i any dst $SERVER_IP &`.
Once tcpdump is running, start pinging the tunnel interface of server node: `ping $SERVER_TUN_IP`. Leave the `ping` process running.
You should see encapsulated packets being captured. For example, notice how the packet destined to the pod network
(172.16.119.192) is encapsulated in a packet that is destined to the node (10.99.224.133) network :
```
20:46:37.409391 IP (tos 0x0, ttl 64, id 51848, offset 0, flags [DF], proto IPIP (4), length 104)
    10.99.224.131 > 10.99.224.133: IP (tos 0x0, ttl 64, id 28044, offset 0, flags [DF], proto ICMP (1), length 84)
    172.16.37.192 > 172.16.119.192: ICMP echo request, id 15784, seq 1, length 64
```
Assuming packets are being sent correctly, we can proceed to verify the server node. Start capturing
packets with the source IP of the client node: `tcpdump -i any -vvv -n src $CLIENT_IP`. You should see the corresponding encapsulated packets being captured:
```
20:50:47.574516 IP (tos 0x0, ttl 61, id 19264, offset 0, flags [DF], proto IPIP (4), length 104)
    10.99.224.131 > 10.99.224.133: IP (tos 0x0, ttl 64, id 10462, offset 0, flags [DF], proto ICMP (1), length 84)
    172.16.37.192 > 172.16.119.192: ICMP echo request, id 19169, seq 2, length 64
```
Assuming packets are arriving to the server node, verify that they are being sent to the tunnel interface.
Capture traffic on the tunnel interface: `tcpdump -vvv -n -i tunl0`. You should see the (unencapsulated) packet
in the capture log:
```
20:52:22.575847 IP (tos 0x0, ttl 64, id 22805, offset 0, flags [DF], proto ICMP (1), length 84)
    172.16.37.192 > 172.16.119.192: ICMP echo request, id 19169, seq 97, length 64
```

If the packets are arriving at the server node, verify that they are finding their way back.

#### Verify server->client connectivity

Capture traffic on the tunnel interface and verify the incoming packets are generating a response: `tcpdump -vvv -n -i tunl0`.
Notice how the first packet has the source IP of the `tunl0` iface of the client node, and the second packet has destination
IP of the `tunl0` iface of the server node.
```
20:52:22.575847 IP (tos 0x0, ttl 64, id 22805, offset 0, flags [DF], proto ICMP (1), length 84)
    172.16.37.192 > 172.16.119.192: ICMP echo request, id 19169, seq 97, length 64
20:52:22.575868 IP (tos 0x0, ttl 64, id 16800, offset 0, flags [none], proto ICMP (1), length 84)
    172.16.119.192 > 172.16.37.192: ICMP echo reply, id 19169, seq 97, length 64
```

Verify that the response packet is being encapsulated and sent out to the right node: `tcpdump -vvv -n -i any dst $CLIENT_IP`.
You should see the encapsulated packet in the capture:
```
20:59:47.594279 IP (tos 0x0, ttl 64, id 14713, offset 0, flags [none], proto IPIP (4), length 104)
    10.99.224.133 > 10.99.224.131: IP (tos 0x0, ttl 64, id 9089, offset 0, flags [none], proto ICMP (1), length 84)
    172.16.119.192 > 172.16.37.192: ICMP echo reply, id 19169, seq 542, length 64
```

Verify that the encapsulated packets are arriving on the client node: `tcpdump -vvv -n -i any src $SERVER_IP`.
You should see the encapsulated packet in the capture:
```
21:02:16.218584 IP (tos 0x0, ttl 61, id 39574, offset 0, flags [none], proto IPIP (4), length 104)
    10.99.224.133 > 10.99.224.131: IP (tos 0x0, ttl 64, id 30288, offset 0, flags [none], proto ICMP (1), length 84)
    172.16.119.192 > 172.16.37.192: ICMP echo reply, id 28537, seq 2, length 64
```

Verify that the packet is arriving on the client node, on the `tunl0` interface: `tcpdump -vvv -n -i tunl0 &`
You should see the ICMP request AND reply packets in the capture:
```
21:04:12.893953 IP (tos 0x0, ttl 64, id 49633, offset 0, flags [DF], proto ICMP (1), length 84)
    172.16.37.192 > 172.16.119.192: ICMP echo request, id 30130, seq 1, length 64
21:04:12.894155 IP (tos 0x0, ttl 64, id 35609, offset 0, flags [none], proto ICMP (1), length 84)
    172.16.119.192 > 172.16.37.192: ICMP echo reply, id 30130, seq 1, length 64
```

#### Verify iptables
Another thing to look into is whether the Calico `iptables` rules are dropping any packets.
The following snippet was captured from a misbehaving node:
```
[root@kismatic-worker ~]# iptables -vL | grep DROP
    0     0 DROP       all  --  any    any     anywhere             anywhere             /* kubernetes firewall for dropping marked packets */ mark match 0x8000/0x8000
    0     0 DROP       all  --  cali+  any     anywhere             anywhere             ctstate INVALID
    0     0 DROP       all  --  any    cali+   anywhere             anywhere             ctstate INVALID
    0     0 DROP       all  --  any    any     anywhere             anywhere             /* From unknown endpoint */
    0     0 DROP       all  --  any    any     anywhere             anywhere             /* From unknown endpoint */
   39  4056 DROP       ipv4 --  any    any     anywhere             anywhere             ! match-set felix-calico-hosts-4 src
```
Notice in the last line that there have been 39 dropped packages. The issue in this specific case was that the
source IP address of the incoming packets was an address different than the one configured with Calico.
