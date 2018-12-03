> A towel, it says, is about the most massively useful thing an interstellar
> hitchhiker can have.
>
> -- <cite>Douglas Adams</cite>

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

### bpftrace

```
bpftrace -l
bpftrace -e 'BEGIN { printf("Hello, World!\n"); }'
```

### bcc tools

TODO: improve examples

```
opensnoop
```

### bcc

TODO: add example

```
```

### docker

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
