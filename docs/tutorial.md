
# Getting Started Tutorial

This guide is intended for those who are new to tracing in Linux. It covers
using the towel container to do various tasks releated to measuring
applications. It is structured around kubernetes and, specifically, GKE since
it is quite easy to get a GKE cluster created.

## Setup kubectl

You will need to have kubectl v1.12 installed. v1.12 is required for the
plugin. If you don't already have it installed check out the [kubernetes
documentation].

You will then need to install the kubectl plugin. If you have the go toolchain
installed you can do so via:

```
go get github.com/jasonkeene/towel/...
```

If you do not have the go toolchain you can download a binary release from the
[releases page].

```
# download the binary
wget -O kubectl-towel <release_binary_url>

# set the permissions bit
chmod +x kubectl-towel

# put it on your path
sudo mv kubectl-towel /usr/local/bin/
```

## Create a GKE Cluster

The first step is to provision a kubernetes cluster. To do so on GKE you will
need a Google Cloud project setup. You will also need to install the [gcloud
SDK]. Once you have the SDK setup you can create a cluster with the following
command:

```
gcloud container clusters create towel-temp \
    --zone us-central1-a \
    --num-nodes 1 \
    --cluster-version 1.11
```

We only need a single node for this tutorial. Creating this cluster will take
a few minutes. Once the cluster is created you can test to see if it works by
running:

```
kubectl version
```

You should see your client version is v1.12 and the server version is v1.11.

## Run a Workload

The workload we will be working with is just a basic nginx container:

```
kubectl create deployment nginx --image nginx
```

This will create the deployment for nginx. You should also expose this via a
LoadBalancer Service to allow you to easily hit it:

```
kubectl expose deployment nginx --port 80 --type LoadBalancer
```

To obtain the public IP address run:

```
kubectl get service nginx --watch
```

Eventually, the `EXTERNAL-IP` will refresh and no longer say `<pending>`. You
can now hit the workload:

```
curl <public_ip>
```

## Run the towel DaemonSet

To deploy the towel DaemonSet simply run:

```
kubectl towel apply
```

You can see if the towel pods are running via:

```
kubectl get pods
```

## Uprobes Demo

Uprobes are my favorite tracing technology. They allow you to trace any
instruction in your application. Lets play around with them using `bpftrace`.

First lets exec onto the towel pod that is on the same node as the nginx pod:

```
kubectl towel exec -l app=nginx
```

Exec'ing onto this pod give us a lot of power. First we are running as root on
the host. This shell also shares the network and pid namespaces with the host.
This gives you a view of everythign that is running on the machine even though
you are exec'ed inside a container. We can use docker to verify that we are on
the host that is running nginx:

```
docker ps | grep k8s_nginx
```

You can also run `pwd` to see that you are in the directory on the host that
represents the root directory for that container. This is just to make things
easier when referencing paths in that container.

First lets use `readelf` to dump the symbol table for nginx:

```
readelf -s usr/sbin/nginx | less
```

There is a lot of stuff here and this information overload can be
intimidating. Lets just search for symbols defined by nginx itself:

```
readelf -s usr/sbin/nginx | grep ngx_ | less
```

There, that is better. So what is all this stuff. The symbol table basically
maps human readable names (`ngx_time_update`) to a place in memory where that
object lives (`0000000000037610`). For tracing purposes we are interested in
locating where functions live in memory so this is useful information. As a
side note, not all binaries have this symbol table. Sometimes a binary can
have this information stripped. This doesn't mean you can not use uprobes,
however, as of the time of writing this you will have to use BCC to trace your
programs and not `bpftrace`.

Lets pick a function to trace, something that is relevant to processing http
requests:

```
readelf -s usr/sbin/nginx | grep ngx_ | grep request
```

The function `ngx_http_process_request` looks interesting. Lets trace it! Run
the following program.

```
bpftrace -e '

uprobe:usr/sbin/nginx:ngx_http_process_request {
    printf("uprobes are awesome!\n");
}

'
```

Now, while that is running, in another window hit your nginx server:

```
curl <public_ip>
```

You can see every time you hit nginx it runs the `ngx_http_process_request`
function and your tracing code gets ran. Neat! Lets do something a little more
advanced. Lets measure the latency of each call to `ngx_http_process_request`.
Here is a program that does just that:

```
bpftrace -e '

uprobe:usr/sbin/nginx:ngx_http_process_request {
    @start[tid] = nsecs;
}

uretprobe:usr/sbin/nginx:ngx_http_process_request
/ @start[tid] /
{
    @ = hist((nsecs - @start[tid])/1000);
    delete(@start[tid]);
}

'
```

This program will run, collect the latencies of each function call, and when
you hit Ctrl+C it will dump out a pretty histogram showing the distribution of
these latencies.

Lets put some load on the nginx server:

```
siege <pubic_ip>
```

Now kill both programs with Ctrl+C and see the wonderful data you just
collected. The amazing thing about this is that we did it without any help
from nginx. nginx did not have to allow us to do this by adding
instrumentation to their code. This is very powerful!

## BCC Tools Demo

The towel pod also includes `libbcc.so` and all the BCC tools. Lets play
around with some of them. Inside the towel pod run:

```
opensnoop -p $(pgrep -n nginx)
```

This will tell you whenever the nginx process opens a file, what the file path
is, if there was an error, and the resulting file descriptor. Try hitting
nginx when `opensnoop` is running:

```
curl <public_ip>
```

That is pretty cool right. Now we can spy on our containers and see what
mischief they are up to!

<!--

TODO: get tcpaccept example working

Lets use `tcpaccept` to monitor the connections nginx is accepting. This one
is a bit tricky. You might remember from earlier that our towel container is
running in the same network namespace as the host. In order to monitor network
activity of a container we need to be in the same network namespace. To do
this we use a program called `nsenter`.

-->

There are plenty of other incredibly useful BCC tools. You can check them out
on the [BCC docs].

## Cleanup

Once you are done with your towel make sure to put it away so it doesn't get
dirty:

```
kubectl towel delete
```

And if you wish to delete the GKE cluster:

```
gcloud container clusters delete towel-temp --zone us-central1-a 
```

[gcloud SDK]: https://cloud.google.com/sdk/
[releases page]: https://github.com/jasonkeene/towel/releases
[kubernetes documentation]: https://kubernetes.io/docs/tasks/tools/install-kubectl/
[BCC docs]: https://github.com/iovisor/bcc#tools
