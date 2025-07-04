# USAGE

DICE is an engine to help orchestrating network measurements.
It was first thought as a tool for Internet surveys, a far broader scope than single host scanning.
However, DICE can be used at many different levels.
It's main purpose is to provide flexibility for as many measurements as you can think of, thus, DICE's modular architecture with plugins.

As an engine to orchestrate measurements, use cases can vary greatly.
Here we include two common examples on how to use DICE.

## Installation

1. __Download DICE's latest release__ using curl or the link on [DICE's github page](https://github.com/RicYaben/dice/releases)

```sh
curl -OL https://github.com/RicYaben/dice/releases/download/latest/dice.linux-amd64.tar.gz
```

2. __Remove any previous DICE installation__ by deleting the /usr/local/dice folder, and extract the new installation

```sh
rm -rf /usr/local/dice && tar -C /usr/local -xzf dice.linux-amd64.tar.gz
```

3. Add dice to your PATH environment variable

Add the following line to your $HOME/.profile file.
__Note:__ this file may be located or named differently depending on your environment and terminal.

```sh
export PATH=$PATH:usr/local/dice/bin
```

4. Verify you have installed DICE.

```sh
dice version
```

## Quickstart

1. __Create a DICE project__ in the current directory

The first time running this command, DICE will setup dependancies, databases, configurations, etc.
Then, DICE will create a `.dice` file with the project configuration.

```sh
dice init d1
```

2. __Run DICE__ with modules or signatures

Using DICE to orchestrate Internet-wide measurements is as simple as using the `scan` command with modules (`-M`) or signatures (`-S`).
Once the measurement is done, you can check the results under the newly created measurement -- check the output logs to find the ID if you did not provide one, -- in the __sources__ directory, or directly from the results `cosmos.db`.

__Note:__ Learn more about DICE measurement results [cosmos databases here](docs/cosmos.md)

```sh
dice scan -M mqtt-anon
```

To list all known to DICE use the following command:

```sh
dice list [-M modules | -S signatures | -P projects] [--all]
```

For more information on DICE's commands and configuration options check out [the list of commands](docs/commands.md).

3. __Query DICE's results__

```sh
dice query 'protocol:mqtt labels>0 mosquitto'
```


## Basic Usage

To start using DICE, we only need to tell the engine which actions we want to use, and which modules or signatures to load.

```bash
# Scan the whole IPv4 using a single signature
dice scan -S router
```

This will output the results to the current directory.
You should see a structure containing (at the very least) a sources folder containing results from the different sources and scanners (e.g., ZGrab2, Censys, Shodan, Greynoise, etc.), and a `cosmos.db` with DICE results.

```text
./
- sources       // 
    - zgrab2    // zgrab2 results
    - zmap      // zmap results
- cosmos.db     // results database
```

However, DICE can do much more, and as with many other tools, the simplest and most common cases, are only the beggining.
Big part of measurements and network analysis is the ability to replay results, and as such, DICE includes commands to ingest data from different sources and classify using a set of modules or full signatures.
This allows for predictive and deterministic results, helping to share and compare processing pipelines and metrics.

```bash
# Classify previously collected zgrab2 records using all signatures 
# that start with "iot-" or "ot-" 
dice classify -S "iot-*,ot-*" --source zgrab2 
```

This command shows how we would classify previously-collected results with a set of signatures.

## Advanced Usage

We have covered the tip of the iceberg, simple scans, simple classifications.
Here we cover DICE more in depth, and showcase other use-cases more interesting for tinkers and those looking to conduct or evaluate complex measurements.

### Concepts

__Signatures:__ DICE register signatures in a shared database located in [DICE's data folder](configuration.md).
This database includes references to the actual signatures and serve as a quick map to validated and known signatures.
Registering a signature requires the referenced modules to be accessible, as DICE will read some values to determine relationships and dependancies.
Those modules must be available at runtime, and are evaluated against the hash value when they were first registered.
If the referenced modules have changed, DICE will throw an error.
This means that making changes to modules or signatures will require users to update the database, or telling DICE to start from an empty database. [Learn more about signatures here](signatures.md)

```bash
// Update signatures database
dice update -S
// Use without database
dice scan -S iot --no-cache
```

__Projects:__ To organize measurements and studies, we can create DICE projects.
Using projects we can maintain the state for scans, share a common configuration for each project, etc.
We view projects as the root for studies under a common set of conditions, e.g., a study that runs for the next 6 months where we will conduct recurrent scans every 2 weeks with the same parameters.

__Components:__

__Modules:__

__Cosmos:__