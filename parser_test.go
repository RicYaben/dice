package dice

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
)

type actionParserTester struct {
	line   string
	action *action
}

func (t *actionParserTester) runTest(test *testing.T, name string) {
	p := &actionParser{}

	action, err := p.Parse(t.line)
	if err != nil {
		test.Errorf("[%s] failed to parse action: %v", name, err)
		return
	}

	if reflect.DeepEqual(action, t.action) {
		test.Errorf("[%s] expected %v, got %v", name, t.action, action)
		return
	}
}

var tests = map[string]*actionParserTester{
	"rule": {
		line: `rule module "router.lua" (sid: tr064-sig; func: tr065;)`,
		action: &action{
			Type:    "rule",
			Headers: []string{"module", `"router.lua"`},
			Options: map[string]string{
				"sid":  "tr064-sig",
				"func": "tr065",
			},
		},
	},
	"probe": {
		line: `probe ftp [21,2121] (flags: authtls true, verbose true;)`,
		action: &action{
			Type:    "probe",
			Headers: []string{"ftp", "[21,2121]"},
			Options: map[string]string{
				"flags": "authtls true, verbose true",
			},
		},
	},
	"probe-simple": {
		line: `probe upnp 1900`,
		action: &action{
			Type:    "probe",
			Headers: []string{"upnp", "1900"},
			Options: map[string]string{},
		},
	},
	"scan": {
		line: `scan tcp (rules: a,b,c; map: template "upnp-location.template";)`,
		action: &action{
			Type:    "scan",
			Headers: []string{"tcp", ""},
			Options: map[string]string{
				"rules": "a,b,c",
				"map":   `template "upnp-location.template"`,
			},
		},
	},
}

func TestActionParser(t *testing.T) {
	for tname, cfg := range tests {
		cfg.runTest(t, tname)
	}
}

type signatureParserTester struct {
	signature string
}

var sigtests = map[string]*signatureParserTester{
	"simple": {
		signature: `
// This is a test rule
// sid -> string ID
probe ssh [22,2222] (flags: userauth true; sid: ssh;)
probe upnp 1900 (sid: upnp;)
rule module "upnp.lua" (sid: upnp-sig;)

// Define scans
scan udp "probes/upnp_1900.pkt" (probes: ssh;)
// scan udp "probes/coap_5683.pkt" (probes: coap;)
scan tcp synack (rules: upnp-sig; probes: upnp,ssh; map: template "upnp-location.template";)
`,
	},
}

func (t *signatureParserTester) runTest(test *testing.T, name string) {
	p := makeSignatureParser()
	r := bufio.NewReader(strings.NewReader(t.signature))

	var sig Signature
	if err := p.Parse(&sig, r); err != nil {
		test.Errorf("[%s] failed to parse signature: %v", name, err)
	}
}

func TestSignatureParser(t *testing.T) {
	for tname, cfg := range sigtests {
		cfg.runTest(t, tname)
	}
}
