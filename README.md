# responsefilter

## Name

*responsefilter* - filters DNS responses based on FQDN and IP CIDR blocklists.

## Description

The responsefilter plugin inspects DNS responses from upstream servers and blocks responses where the returned IP address matches a configured blocklist for specific domains. This helps protect against DNS spoofing and malicious responses.

## Syntax

```
responsefilter {
    block DOMAIN CIDR [CIDR...]
}
```

* **DOMAIN** - the domain name to apply the filter to (supports subdomains)
* **CIDR** - one or more IP CIDR ranges to block for this domain

**IMPORTANT:** The `responsefilter` directive must be placed **before** the `forward` directive in your Corefile so it can intercept responses from upstream servers.

## Examples

Block specific IP ranges for a domain:

```
.:53 {
    responsefilter {
        block example.com 10.1.1.0/24
    }
    forward . 8.8.8.8
}
```

Block multiple CIDR ranges for multiple domains:

```
.:53 {
    responsefilter {
        block abc.com 10.1.1.0/24 192.168.0.0/16
        block xyz.com 172.16.0.0/12
    }
    forward . 8.8.8.8
}
```

## Behavior

When a DNS response contains an A or AAAA record with an IP that matches a blocked CIDR for the queried domain:
- The response is dropped
- A REFUSED response is returned to the client
- The original response is not sent

This works for both A (IPv4) and AAAA (IPv6) records.

## Installation

### Option 1: Use as external plugin

Add the plugin to your CoreDNS `plugin.cfg`:

```
responsefilter:github.com/isovalent/responsefilter
```

Then rebuild CoreDNS:

```bash
go generate
go build
```

### Option 2: Build from this repository

Clone the CoreDNS repository and add the plugin to `plugin.cfg`. To build:

```bash
# Generate plugin registry
go generate

# Build for ARM64 (for kind on Apple Silicon)
podman buildx build --platform linux/arm64 -t localhost/coredns:v3 --load .

# Or build for AMD64
podman buildx build --platform linux/amd64 -t localhost/coredns:v3 --load .
```

### Load image into kind cluster

```bash
# Save the image to a tar file
podman save localhost/coredns:v3 -o /tmp/coredns-v3.tar

# Load into all kind nodes
for node in kind-control-plane kind-worker kind-worker2; do
  echo "Loading image into $node..."
  podman cp /tmp/coredns-v3.tar $node:/tmp/coredns-v3.tar
  podman exec $node ctr -n k8s.io images import /tmp/coredns-v3.tar
  podman exec $node rm /tmp/coredns-v3.tar
done
```

### Restart CoreDNS

```bash
# Set your kubeconfig
export KUBECONFIG=/path/to/kubeconfig

# Delete CoreDNS pods to restart with new image
kubectl delete pods -n kube-system -l k8s-app=kube-dns

# Wait for pods to be ready
kubectl wait --for=condition=ready pod -n kube-system -l k8s-app=kube-dns --timeout=60s
```

### Testing

Test the plugin by querying a blocked domain from within the cluster:

```bash
# Create a test pod
kubectl run -it --rm debug --image=busybox --restart=Never -- sh

# Inside the pod, test DNS queries
# This should return a normal response (if abc.com doesn't resolve to 10.x.x.x)
nslookup abc.com

# If abc.com resolves to an IP in 10.1.1.0/24, you should see:
# ** server can't find abc.com: REFUSED
```

Or test from outside the cluster:

```bash
# Get CoreDNS service IP
kubectl get svc -n kube-system kube-dns

# Query using dig or nslookup
dig @<coredns-ip> abc.com

# Check for REFUSED status in blocked cases
```

### Verify plugin is loaded

Check CoreDNS logs to confirm the plugin loaded successfully:

```bash
kubectl logs -n kube-system -l k8s-app=kube-dns --tail=20
```

You should see CoreDNS start without errors. The configuration SHA will change when you update the ConfigMap.
