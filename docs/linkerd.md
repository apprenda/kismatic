# linkerd

The [linkerd](https://linkerd.io) service mesh is a process that runs on each
node of your Kubernetes cluster, manages the communication between pods,
and provides:

* [load balancing](https://linkerd.io/features/load-balancing/)
* [circuit-breaking](https://linkerd.io/features/circuit-breaking/)
* [service discovery](https://linkerd.io/features/service-discovery/)
* [dynamic request routing](https://linkerd.io/features/routing/)
* [TLS](https://linkerd.io/features/tls/)
* [and more](https://linkerd.io/features)

## Deploying linkerd on Kismatic

Once you have a working Kismatic cluster, linkerd can be deployed with one
command:

```
kubectl --kubeconfig generated/kubeconfig apply -f ansible/roles/addon-linkerd/templates/linkerd.yml
```

## Using linkerd

To take advantage of the linkerd service mesh your apps must do two things.
First, your app must send requests through linkerd. For many applications, this
can be done without code changes by
[using linkerd as an HTTP proxy](https://linkerd.io/features/http-proxy/).
For example, Go, C, Ruby, Perl, and Python applications can set the `http_proxy`
environment variable to direct all HTTP calls through linkerd.  You should set
the environment variable to the node name on port 4140.  This can easily be
accomplished using the
[downward API](http://kubernetes.io/docs/user-guide/downward-api/), like this:

```
env:
- name: NODE_NAME
  valueFrom:
    fieldRef:
      fieldPath: spec.nodeName
command:
- "/bin/bash"
- "-c"
- "HTTP_PROXY=$(NODE_NAME):4140 python hello.py"
```

Second, the Kubernetes service object for your app must define a port named
`http` in the default namespace where the app will accept requests (these
defaults can be changed in the linkerd.yml).

Applications can then make requests to `http://hello` and the request will
be routed to the hello service through the linkerd service mesh.

## More Info

For more info about linkerd as a service mesh for Kubernetes, see
[this blog series](https://blog.buoyant.io/2016/10/04/a-service-mesh-for-kubernetes-part-i-top-line-service-metrics/)
or visit [linkerd.io](https://linkerd.io).
