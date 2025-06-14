# USAGE

#### Diagram
Creates a diagram for a deployment

```text
--output,-o <filepath>
--signatures,-s <comma separated list of globs>
--config,-c <filepath>
```

#### Replay
Replay a deployment.

```text
--input,-i <filepath>
--output,-o <filepath>
--signatures,-s <comma separated list of globs>
--config,-c <filepath>
```

#### Deploy
Run a list of signature deployments. Including scans

```text
--signatures,-s,(default) <comma separated list of globs>
--config,-c <filepath> 
```

## Syntax

### Headers
```text
probe,scan,rule
```

#### probe

#### scan
```text
@targets, @probe: <tool (optional)> <l4> <probe>
tool: zmap, zgrab2, nmap
l4: tcp, upd, icmp
probe: multi, http, webproxy, coap, mqtt...
```

#### rule