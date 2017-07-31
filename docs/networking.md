# Networking
KET uses CNI as the solution for cluster networking, and it supports multiple
CNI providers out of the box. Choosing the CNI provider will depend on your specific
infrastructure and application needs. 

The CNI provider is exposed as an Add-On in the KET plan file via the 
`add_ons.cni.provider` field. See the [CNI Add-On reference documentation](ADD_ONS.md#cni)
for more information on how to configure CNI.


### Default CNI Provider
The default CNI provider used by KET is Calico for the following reasons:
* Strong network policy support: it supports granular network policy for true SDN patterns at the Pod level.
* Based on routable, layer 2/3 primitives, instead of overlays/encapsulation, making any network debugging much easier and predictable
* Supported by a commercial entity with years of operational experience

However, other CNI plugins provide features and compatibility that may be more appropriate for your particular cloud or architecture requirements.

### CNI Provider Comparison
The following table attempts to list key characteristics of each supported implementation.

|  | [Calico](https://www.projectcalico.org/) | [Weave](https://www.weave.works/oss/net/) | [Contiv](https://contiv.github.io/) |
|---|--------|-------|--------|
| Data Path Technique | L3 with BGP Peering or IPIP Encapsulation | UDP Encapsulation | VXLAN |
| Requires etcd cluster | Yes | No | Yes |
| Multicast Support | No | Yes | Yes |
| Ingress Policy | Yes | Yes | Yes<sup>1</sup> |
| Egress Policy | Yes | No | Yes |
| Can Encrypt Traffic | No | Yes | No |

<sup>1. Contiv does not support the Kubernetes Network Policy API. It uses a custom mechanism for applying policy.</sup>

## Calico Notes
Calicoctl is the command-line utility for managing the Calico network.

If you need to troubleshoot calico, using calicoctl will be useful. This is
a quick command that you can use to run calicoctl:
```
docker run -i \
    --net host \
    -v /etc/kubernetes:/etc/kubernetes \
    -v /etc/calico/calicoctl.cfg:/etc/calico/calicoctl.cfg \
    calico/ctl:v1.1.0
```

Links: 
* Troubleshooting docs: http://docs.projectcalico.org/v2.3/usage/troubleshooting/
* Reference docs: http://docs.projectcalico.org/v2.3/reference/

## Weave Notes

Links:
* How it works: https://www.weave.works/docs/net/latest/concepts/how-it-works/
* Operational Guide: https://www.weave.works/docs/net/latest/operational-guide/
* Troubleshooting: https://www.weave.works/docs/net/latest/troubleshooting/

## Contiv Notes
KET supports Contiv as a "preview", as it is still under active development.

The following are known issues you should be aware of if you choose to install Contiv:
* https://github.com/contiv/netplugin/issues/940
* https://github.com/contiv/netplugin/issues/937
* https://github.com/contiv/netplugin/issues/871
* https://github.com/contiv/netplugin/issues/777
* https://github.com/contiv/netplugin/issues/942

Useful Links:
* Policies: https://contiv.github.io/documents/networking/policies.html
* Admin guide: https://contiv.github.io/documents/admin/index.html
* CLI reference: https://contiv.github.io/documents/reference/netctlcli.html