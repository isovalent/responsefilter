# CoreDNS Response Filter Plugin

A CoreDNS plugin that filters DNS responses based on FQDN and IP CIDR blocklists to protect against DNS spoofing and malicious responses.

authors: pijablon@cisco.com + vibecoding

## Overview

The `responsefilter` plugin inspects DNS responses from upstream servers and blocks responses where the returned IP address matches a configured blocklist for specific domains. When a blocked response is detected, CoreDNS returns a `REFUSED` status instead of the spoofed IP.

## Features

- **FQDN-based filtering**: Apply IP blocklists to specific domains
- **CIDR support**: Block entire IP ranges using CIDR notation
- **Subdomain matching**: Rules apply to subdomains automatically
- **IPv4 and IPv6**: Works with both A and AAAA records
- **Multiple rules**: Configure multiple domain/CIDR combinations

## Installation

### Prerequisites

- Go 1.24 or later
- CoreDNS source code

### Add to CoreDNS

1. Clone your CoreDNS repository
2. Edit `plugin.cfg` and add this line **before** the `forward` plugin:

```
responsefilter:github.com/isovalent/responsefilter
```

3. Rebuild CoreDNS:

```bash
go generate
go build
```

## Configuration

Add the `responsefilter` directive to your Corefile **before** the `forward` directive:

```
.:53 {
    responsefilter {
        block example.com 10.0.0.0/8
        block malicious.net 192.168.0.0/16 172.16.0.0/12
    }
    forward . 8.8.8.8
}
```

### Syntax

```
responsefilter {
    block DOMAIN CIDR [CIDR...]
}
```

- **DOMAIN**: The domain name to filter (supports subdomains)
- **CIDR**: One or more IP CIDR ranges to block for this domain

### Important

The `responsefilter` directive **must** be placed before the `forward` directive in your Corefile so it can intercept responses from upstream servers.

## Examples

### Block specific IP range for a domain

```
.:53 {
    responsefilter {
        block abc.com 10.1.1.0/24
    }
    forward . 8.8.8.8
}
```

### Block multiple CIDR ranges

```
.:53 {
    responsefilter {
        block abc.com 10.1.1.0/24 192.168.0.0/16
        block xyz.com 172.16.0.0/12
    }
    forward . 8.8.8.8
}
```

### Full Kubernetes CoreDNS example

```
.:53 {
    errors
    health {
       lameduck 5s
    }
    ready
    
    kubernetes cluster.local in-addr.arpa ip6.arpa {
       pods insecure
       fallthrough in-addr.arpa ip6.arpa
       ttl 30
    }
    
    responsefilter {
        block abc.com 10.1.1.0/24
        block suspicious-domain.com 10.0.0.0/8
    }
    
    prometheus :9153
    forward . 8.8.8.8 8.8.4.4
    cache 30
    loop
    reload
    loadbalance
}
```

## How It Works

1. DNS query arrives at CoreDNS
2. Query is forwarded to upstream DNS server
3. Response is intercepted by `responsefilter`
4. Plugin checks if any A/AAAA records match blocked FQDN + CIDR combinations
5. If match found: Returns `REFUSED` to client
6. If no match: Passes response through normally

## Testing

### In Kubernetes

```bash
# Create a test pod
kubectl run -it --rm debug --image=busybox --restart=Never -- sh

# Inside the pod, test DNS
nslookup abc.com

# If blocked, you'll see:
# ** server can't find abc.com: REFUSED
```

### Using dig

```bash
# Get CoreDNS service IP
kubectl get svc -n kube-system kube-dns

# Query the domain
dig @<coredns-ip> abc.com

# Check for REFUSED status in response
```

## Deployment to Kubernetes (kind)

### Build CoreDNS image

```bash
# For ARM64 (Apple Silicon)
podman buildx build --platform linux/arm64 -t localhost/coredns:v3 --load .

# For AMD64
podman buildx build --platform linux/amd64 -t localhost/coredns:v3 --load .
```

### Load into kind cluster

```bash
# Save image
podman save localhost/coredns:v3 -o /tmp/coredns-v3.tar

# Load into all kind nodes
for node in kind-control-plane kind-worker kind-worker2; do
  podman cp /tmp/coredns-v3.tar $node:/tmp/coredns-v3.tar
  podman exec $node ctr -n k8s.io images import /tmp/coredns-v3.tar
  podman exec $node rm /tmp/coredns-v3.tar
done
```

### Update CoreDNS ConfigMap

```bash
kubectl edit configmap coredns -n kube-system
```

Add the `responsefilter` block before `forward`.

### Restart CoreDNS

```bash
kubectl delete pods -n kube-system -l k8s-app=kube-dns
kubectl wait --for=condition=ready pod -n kube-system -l k8s-app=kube-dns --timeout=60s
```

## Troubleshooting

### Check CoreDNS logs

```bash
kubectl logs -n kube-system -l k8s-app=kube-dns --tail=50
```

### Verify plugin loaded

CoreDNS should start without errors. Look for the configuration SHA in logs.

### Architecture mismatch

If CoreDNS crashes with segmentation fault, ensure the image architecture matches your nodes:

```bash
# Check node architecture
kubectl get nodes -o wide

# Build for correct architecture (arm64 or amd64)
```

## License

Apache License 2.0 - See [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please open an issue or pull request.

## Support

For issues and questions, please open a GitHub issue at https://github.com/isovalent/responsefilter
