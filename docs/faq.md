# FAQ

### Q: Is DICE replacing other network scanners? E.g., Nmap, ZMap, ZGrab2, etc.

__No!__ DICE is not a network scanner :) instead, DICE uses your favourite scanners under the hood to conduct all shorts of measurements.

### Q: What makes DICE different from services like Shodan, Censys, etc.?

Simply put, DICE hands you the control over measurements.

DICE is an extensible classification engine that goes beyond collecting and serving data.
With DICE, we can make and test hypothesis, share experiments, and conduct longitudinal studies.
There are many reasons why this is important:

* Control over the data
* Custom probes
* On demand measurements
* Logic-driven measurements
* Compare and contrast sources
* ...

Moreover, DICE ingests and uses data to identify and classify devices with different levels of granularity, and help to contrast information from different sources.
Lastly, DICE helps to monitor behavioral changes in remote networks; comparing results between measurements and experiments is paramount to understand how systems evolve.

### Q: Can I use DICE to scan a network without classifying the results? (just orchestrate scans)

__Yes**__
(but you probably shouldn't).
Technically, the answer is yes; DICE is very flexible and you can probably find a way to do this on a flat structure of dummy modules.

```bash
// Scan using a flat structure of modules that just suggest scans
dice scan -M mqtt-topics,opcua-nodes,telnet-brute

// Example module: telnet-brute
&dice.Classifier{
    Require: &dice.Scan{
        Scanner: 'zgrab2', 
        Arguments: map[string]any{
            "port": 23,
            "probe": "telnet",
            "flags": map[string]any{
                ...
            },
        },
    },
}
```

### Q: Can you give me a few use cases for  DICE?

* Monitoring the evolution of security awareness around the globe (e.g., measuring digital divide)
* Measuring the population of vulnerable critical infrastructure devices facing the Internet
* Evaluating the security posture at particular organizations
* Monitoring outages in areas of conflict
* Overcoming Internet churn to find vulnerable IoT devices in residential networks
* Identifying sudden changes in known static devices (e.g., botnet infections)
* ...

### Q: Can I use DICE to classify previously gathered results?

__Yes.__
DICE can ingest results from a variety of sources, check out the current list of [compatible sources](compatibility.md).

### Q: How can I reference DICE in my publications?

You should use the following reference.
Thank you for helping DICE to improve!

```latex
@misc{dice,
    title: "DICE: Device Identification and Classification Engine",
    authors: "Yaben, Ricardo and Vasilomanolakis, Emmanouil",
    url: "https://github.com/RicYaben/dice.git",
}
```

### Q: Is there a list of open access research publications using DICE?

__Yes.__
We maintain a [separate repository](https://github.com/RicYaben/dice-publications.git) with all the publications where we used DICE, including presentation slides, posters, etc.
