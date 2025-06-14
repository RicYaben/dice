<div id="banner">
</div>

#### Device Identification and Classification Engine (DICE)
# DICE

DICE is an engine to describe and deploy complex scans efficiently with top-level syntax.
Using a directed graph structure, DICE propagates scan records to dependent signatures as they meet criteria.

We define signatures as functions, algorithms or models that flag hosts and records.
This loose definition makes signatures elastic, where a signature can vary from a string matching function, to deep neural networks.

The output of DICE is a labelled dataset and summary reports with the findings.

## Usage

```bash
# Scan for Web proxies
dice deploy -o results --signatures webproxy
# Label a previous dataset for routers
dice describe -i results --signatures router
# Show a diagram of the signatures combined
dice preview --signatures "webproxy,router"
```

We identify the following use-cases for DICE.
First (`deploy`), to orchestrate complex scans with dependencies that require making dynamic decissions.
Second (`describe`), to identify and classify host behaviors based on previously gathered information.
Lastly (`preview`), sharing a common language to describe scans.
In summary, DICE ships with the following commands:

| Command | Description |
|---|---|
| `deploy` | mainly aimed at targetted scans with one or more signatures. May start dynamic scans |
| `describe` | labels previously generated datasets and summarizes results |
| `preview` | Preview signatures |

## Signatures

DICE conceptualizes signatures as plugin elements communicating over gRPC calls taking the latest record added to a host, and returning labels and annotations.
Signatures run in sandboxed environments when needed without access to the system and without privileges.
We provide an SDK to write signatures for the following languages:

| Language | Version |
|--|--|
| python | <= 3.12 |
| lua | any |
| ruby | any |
| perl | any |
| R | any |
| go | <= 1.22 |

Signatures can be published as containers.
Signatures that require access to remote resources are marked as `unsafe`. 