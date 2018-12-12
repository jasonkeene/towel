> A towel, it says, is about the most massively useful thing an interstellar
> hitchhiker can have.
>
> -- <cite>Douglas Adams</cite>

## Getting Started Tutorial

If you are new to linux tracing I highly recommend you read our [Getting
Started Tutorial]. It will walk you through step-by-step how to get started
tracing your programs running in a GKE cluster.

## Installation

```
go get github.com/jasonkeene/towel/...
```

## Usage

```
# apply the towel daemonset
kubectl towel apply

# exec onto the towel for a given pod
kubectl towel exec nginx
kubectl towel exec -l app=nginx --field-selector spec.nodeName=node-1

# delete the towel daemonset
kubectl towel delete
```

## Things You Can Do with the Towel

### Run bpftrace

```
bpftrace -l
bpftrace -e 'BEGIN { printf("Hello, World!\n"); }'
```

### Run BCC Tools

```
opensnoop
tcpaccept
tcpconnect
```

### Write Your Own BCC Tools

```
python -c '

from bcc import BPF
BPF(text=r"""
int kprobe__sys_clone(void *ctx)
{
    bpf_trace_printk("Hello, World!\n");
    return 0;
}
""").trace_print()

'
```

### Talk to the Docker Daemon

```
docker ps
```

## Downloading GKE Kernel Source

You might run into situations where you need the kernel source. If you are
running the chromium OS image on GKE you can use the following:

```
download-chromium-os-kernel-source
```

It will print out the environment variables you need to export to allow BCC
and bpftrace to know where to look for the sources.

[Getting Started Tutorial]: docs/tutorial.md
