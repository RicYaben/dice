# SIGNATURES

DICE orchestrates measurements using configuration files indicating relationships and dependancies between the steps to follow during the measurement.
These configuration files are so-called __signatures__.
The term _signature_ should sound familiar to those who have ever used Intrusion Detection Systems (IDSs) such as Snort or Suricata, and even -- in some way, -- Iptables.

DICE syntaxt to compose signatures is simplistic in nature, referencing one or more modules and assigning dependencies when needed.
Take the following signature as an example to scan, identify and classify routers facing the Internet.

```lua
-- scan for UPnP services
cls upnp

-- check whether the device responds to router probes,
-- if it has a http service mentioned in the upnp banner
-- and if there is other sensitive information being disclossed
cls router (cls: upnp)

-- check if the router is from zyxel, and the issues related to it
-- e.g.: CVEs,  
cls zyxel (cls: router)
```

Then, we register the signature and invoke a new scan using this signature, which we named `routers.dice`.

```sh
// add the signature to the database
dice add -S routers

// scan the whole IPv4 space using this signature
dice scan -S routers
```

As you can see, there is no need to specify which scanner to use, identifiers, or any other parameters.
When registering a new signature, DICE makes sure there are no conflicts, such as, circular dependancies (i.e., closed loops), the modules can be located and loaded, etc.
In addition, DICE will check which calls the module makes to figure which scanners and identifiers are needed.
Take the following example of a module that works with ZGrab data.

```go
import dice

func process(d map[string]any) error {...}

func Module() *dice.Classifier {
    return &dice.Classifier{
        Name: "example",
        Description: "something about the module",
        Query: dice.Fingerprint{Source: "zmap", Protocol: "upnp"},
        Run: func(c *dice.Classifier, args []string) error {
            return process(c.Data)
        },
    }
}
    
```

This example shows how DICE pre-load modules; modules including explicit queries indicating which data they require.
These queires allow DICE to know which identifier modules are needed.
Similarly, identifier modules indicate which data sources they can identify, allowing data to load necessary scanners.
The more explicit the query, the fewer modules will be loaded.
The most explicit query DICE accepts is calling objects for their unique identifier. For example:

```bash
Query: dice.Fingerprint{Name: "zmap-upnp"}
```

Note this is only an example, and most identifiers will be able to handle data from a variety of sources.

To find more about identifiers and other modules, check the [list of DICE commands](commands.md).

```bash
// The complete DICE documentation 
dice docs
// Find information about a signature 
dice help -M example-module
```

## Advanced signatures

Complex measurements may require tinkering with modules and parameters.
Often, we may need to change the default behavior of one or more modules to achieve something in particular.
For this, DICE allows passing arbitrary arguments to modules.
In the following example, we change the default behavior of a single module.

```bash
cls upnp (args: 'port=100 transport="tcp"')
cls router (cls: upnp args: '')
```

Other common parameters are often set in DICE's configuration file or passed as arguments when calling DICE.

```bash
dice -S routers --module-args 'user-agent="DICE"'
```

While these expressions are sufficient to cover the vast majority of measurements, there may be use-cases we have not considered.
If you find yourself limited by DICE, [get in contact with the team](contact.md).
