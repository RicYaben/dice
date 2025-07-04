<p align="center">

![Header](./docs/logo/banner.png)

</p>

<p align="center">
</p>

DICE is an engine to orchestrate Internet measurements.

The output of DICE is a labelled dataset and summary reports with the findings.

## Usage

```bash
# Scan the whole Internet with IoT and OT signatures
dice scan --signatures iot,ot
# Label a previous dataset
dice classify -i results --signatures iot,ot
# Show a diagram of the signatures combined
dice preview --signatures webproxy,router
# query a database
dice query 'port:104 protocol:dicom tag:healthcare ORTHANC' cosmos.db
```

We identify the following use-cases for DICE.
First (`deploy`), to orchestrate complex scans with dependencies that require making dynamic decissions.
Second (`describe`), to identify and classify host behaviors based on previously gathered information.
Lastly (`preview`), sharing a common language to describe scans.
In summary, DICE ships with the following commands:

| Command | Description |
|---|---|
| `scan` | mainly aimed at targetted scans with one or more signatures |
| `classify` | labels previously generated datasets and summarizes results |
| `preview` | Preview signatures |

## Signatures

We provide an SDK to write signatures for the following languages:

| Language | Version |
|--|--|
| python | 3.12 |
| lua | any |
| ruby | any |
| perl | any |
| R | any |
| go | 1.22 |