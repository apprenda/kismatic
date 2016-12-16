# Basic Ingress for Kismatic

The very first question anybody has upon installing Kubernetes is “how do I access my workloads?”

With overlay or policy-enforcing networking in play, this question becomes even more imperative. Answers such as “join the Kubernetes Pod network,” “SSH to one of the nodes” or “Use a NodePort service” all have flaws from a security and usability perspective.


For Layer 7 (HTTP/S) services, the best answer available is “use an Ingress.” [Ingress](http://kubernetes.io/docs/user-guide/ingress/) allows an HTTP server to be used along with port and path mapping to present an http service. Ingress nodes sit between the pod network and the local network, brokering HTTPx traffic from the local network into the pod network, with the ability to terminate TLS.

For Kismatic to support Ingress, we will present a new class of node: `ingress`  
This node will contain:
* `kubelet` and be part of the kubernetes cluster, by default the **kubelet will be unschedulable** on the ingress nodes
* The certificates required to communicate with the kubernetes cluster
* a [default backend](https://github.com/kubernetes/contrib/tree/master/404-server) required for the ingress controller
  * The backend will run as a [Deamon Set](http://kubernetes.io/docs/admin/daemons/) on the `ingress` nodes, with a kubernetes service fronting it
* an [Nginx Ingress Controller](https://github.com/kubernetes/contrib/tree/master/ingress/controllers/nginx) that will listen on ports **80** and **443**
  * The controller will run as a [Deamon Set](http://kubernetes.io/docs/admin/daemons/) on the `ingress` nodes with `hostPort: 80` and `hostPort: 443`
  * The controller will run with `hostNetwork: true`, see [issue 23920](https://github.com/kubernetes/kubernetes/issues/23920)
  * The controller has a `/healthz` endpoint that will return a `200` status when its alive
  * The controller will respond with a `404` when a requested endpoint is not mapped with a ingress resource

For HA configurations it is recommended to have **2 or more** ingress nodes and a load balancer configured with the nodes' addresses, using the `/healthz` endpoint to maintain a list of healthy nodes

### Plan File changes

```
...
worker:
  expected_count: 3
  nodes:
  - host: node1.somehost.com
    ip: 8.8.8.1
    internalip: 8.8.8.1
  - host: node2.somehost.com
    ip: 8.8.8.2
    internalip: 8.8.8.2
  - host: node3.somehost.com
    ip: 8.8.8.3
    internalip: 8.8.8.3
ingress:
  expected_count: 2
  nodes:
  - host: node4.somehost.com
    ip: 8.8.8.4
    internalip: 8.8.8.4
  - host: node1.somehost.com
    ip: 8.8.8.1
    internalip: 8.8.8.1
```

To support new node types an optional `ingress` section will be added    
When an ingress section is not provided, the ingress controller will NOT be setup  
`ingress` can have 1 or more nodes, these nodes can be unique from the other roles or can be shared
* On an `ingress` node the kubelet will be **unschedulable**, ie. `node4.somehost.com` from the example
* If the node is only shared with `etcd` or/and `master ` the kubelet will be **unschedulable**
* If the `ingress` node is also a `worker` the kubelet will be **schedulable**, ie. `node1.somehost.com` from the example

### Example Ingress Resources
Assumptions:
* at least 1 `ingress` node was provided when setting up the cluster
* a service named `echoserver` with `port: 80` is running in the cluster
* replace `mydomain.com` with your actual domain
* you configured `mydomain.com` to resolve to your ingress node(s)

To expose via HTTP on port 80 of the ingress node, `kubectl apply`:
```
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: echoserver
  annotations:
    kubernetes.io/ingress.class: "nginx"
spec:
  rules:
  - host: mydomain.com
    http:
      paths:
      - path: /echoserver
        backend:
          serviceName: echoserver
          servicePort: 80

```

To expose via HTTPS on port 443 of the ingress node, `kubectl apply`:
```
echo "
apiVersion: v1
kind: Secret
metadata:
  namespace: echoserver
  name: mydomain.com-tls
data:
  tls.crt: `base64 /tmp/tls.crt`
  tls.key: `base64 /tmp/tls.key`
" | kubectl create -f -
```
where `tmp/tls.crt` and `/tmp/tls.key` are the certificates generated with the `mydomain.com` CN
```
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: echoserver
  annotations:
    kubernetes.io/ingress.class: "nginx"
spec:
  tls:
  - hosts:
    -  mydomain.com
    secretName:  mydomain.com-tls
  rules:
  - host: mydomain.com
    http:
      paths:
      - path: /echoserver
        backend:
          serviceName: echoserver
          servicePort: 80
```

After running the above, your service will be accessible vi `http://mydomain.com/echoserver` and `https://mydomain.com/echoserver`

### Out of Scope
* Integrating with any cloud provider for Load Balance functionality - this enhancement should be added along with the kubernetes API server HA
* Automatic HTTPs cert generation, the domain owner will either already have certificates or an existing workflow to create new certificates
  * [kube-lego](https://github.com/jetstack/kube-lego) was evaluated as a possible integration point with Let's Encrypt but the domain needs to be already configured with [ACME](https://letsencrypt.github.io/acme-spec/) to function
* Any functionality after setting up the ingress controller, the user of the cluster will still need create ingress resources
