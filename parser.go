package dice

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/dice/pkg/parser"
)

var (
	ErrEmptyAction     = fmt.Errorf("empty action")
	ErrMalformedAction = fmt.Errorf("malformed action")
)

var (
	ActionExp  = regexp.MustCompile(`^(\w+)\s+([^\(\)]+)?\s*(?:\((.*)\))?$`)
	OptionsExp = regexp.MustCompile(`(\w+):\s*([^;]+)`)
)

type mapper struct {
	mode   string
	source string
}

type scanWrapper struct {
	rules  []string
	probes []string
	scan   Scan
}

type scanHeaders struct {
	// L4 protocol
	protocol string
	// L4 packet to send. E.g.: "synack".
	// If empty or none, throw directly the results to the L6 scanner
	module string
}

type scanOptions struct {
	// probes to use during L6 scan
	probes []string
	rules  []string
	// mapper to convert a host and record into a valid target.
	// Note: it may be better to use a function though.
	// A new syntax to convert a record and host into a valid
	// target is not the best solution.
	mapper Mapper
}
type scanBuilder struct {
	headers scanHeaders
	options scanOptions
}

func newScanBuilder() *scanBuilder {
	b := &scanBuilder{}
	b.reset()
	return b
}

func (b *scanBuilder) reset() {
	b.headers = scanHeaders{
		"tcp",
		"skip",
	}

	b.options = scanOptions{
		mapper: Mapper{
			Mode:   "direct",
			Source: "default",
		},
	}
}

// Note: we don't check here. The module decides whether this is valid
func (b *scanBuilder) setMapper(m string) *scanBuilder {
	mFields := strings.Fields(m)
	b.options.mapper.Mode = mFields[0]
	b.options.mapper.Source = mFields[1]
	return b
}

func (b *scanBuilder) SetHeaders(headers []string) error {
	checkHeader := func(arr []string, index int, defaultVal string) string {
		if index <= len(arr) && arr[index] != "" {
			return arr[index]
		}
		return defaultVal
	}

	h := b.headers
	b.headers.protocol = checkHeader(headers, 0, h.protocol)
	b.headers.module = checkHeader(headers, 1, h.module)
	return nil
}

func (b *scanBuilder) SetOptions(ops map[string]string) error {
	var burnt []string
	for key, val := range ops {
		if slices.Contains(burnt, key) {
			return fmt.Errorf("found duplicated option: %v", key)
		}
		burnt = append(burnt, key)

		switch key {
		case "map":
			b.setMapper(val)
		case "probes":
			b.options.probes = strings.Split(val, ",")
		case "rules":
			b.options.rules = strings.Split(val, ",")
		default:
			return fmt.Errorf("scan option not allowed: %s", key)
		}
	}
	return nil
}

func (b *scanBuilder) Wrapper(ac *action) (scanWrapper, error) {
	defer b.reset()
	var wrapper scanWrapper
	if err := b.SetHeaders(ac.Headers); err != nil {
		return wrapper, fmt.Errorf("failed to set rule headers: %v", err)
	}

	if err := b.SetOptions(ac.Options); err != nil {
		return wrapper, fmt.Errorf("failed to set rule options: %v", err)
	}

	wrapper.probes = b.options.probes
	wrapper.rules = b.options.rules
	wrapper.scan = Scan{
		Protocol: b.headers.protocol,
		Module:   b.headers.module,
		Mapper:   b.options.mapper,
	}
	return wrapper, nil
}

type RuleMode string

const (
	R_MODULE RuleMode = "module"
	R_SOURCE RuleMode = "source"
)

type ruleWrapper struct {
	rules []string
	rule  Rule
}

type ruleHeaders struct {
	mode   string
	source string
}

type ruleOptions struct {
	sid   string
	role  string
	rules []string
}

type ruleBuilder struct {
	modes []RuleMode

	headers ruleHeaders
	options ruleOptions
}

func newRuleBuilder() *ruleBuilder {
	b := &ruleBuilder{
		modes: []RuleMode{R_MODULE, R_SOURCE},
	}
	b.reset()
	return b
}

func (b *ruleBuilder) reset() {
	b.headers = ruleHeaders{
		"module",
		"fake",
	}
	b.options = ruleOptions{}
}

func (b *ruleBuilder) setMode(mode string) error {
	if !slices.Contains(b.modes, RuleMode(mode)) {
		return fmt.Errorf("rule mode not allowed %s. Options: %v", mode, b.modes)
	}
	b.headers.mode = mode
	return nil
}

func (b *ruleBuilder) SetHeaders(headers []string) error {
	if len(headers) != 2 {
		return fmt.Errorf("headers must include mode and source")
	}

	if err := b.setMode(headers[0]); err != nil {
		return err
	}
	b.headers.source = headers[1]
	return nil
}

func (b *ruleBuilder) SetOptions(options map[string]string) error {
	var burnt []string

	for key, val := range options {
		if slices.Contains(burnt, key) {
			return fmt.Errorf("found duplicated option: %v", key)
		}
		burnt = append(burnt, key)

		// TODO: Convert this into an unmarshaller. Right now we don't
		// really know if there are more options. The unmarshaller
		// should take an action and apply some functions to each keyword
		switch key {
		case "sid":
			b.options.sid = val
		case "rules":
			b.options.rules = strings.Split(val, ",")
		case "role":
			b.options.role = val
		case "func":
			// TODO FIXME: implement this
			// ignore for now
		default:
			return fmt.Errorf("rule option not allowed: %s", key)
		}
	}

	return nil
}

func (b *ruleBuilder) Wrapper(ac *action) (ruleWrapper, error) {
	defer b.reset()
	var wrapper ruleWrapper

	if err := b.SetHeaders(ac.Headers); err != nil {
		return wrapper, fmt.Errorf("failed to set rule headers: %v", err)
	}

	if err := b.SetOptions(ac.Options); err != nil {
		return wrapper, fmt.Errorf("failed to set rule options: %v", err)
	}

	wrapper.rules = b.options.rules
	wrapper.rule = Rule{
		Sid:    b.options.sid,
		Mode:   b.headers.mode,
		Source: b.headers.source,
		Role:   b.options.role,
	}
	return wrapper, nil
}

type probeHeaders struct {
	protocol string
	ports    []uint16
}

type probeOptions struct {
	sid   string
	flags map[string]string
}

type probeBuilder struct {
	headers probeHeaders
	options probeOptions
}

func newProbeBuilder() *probeBuilder {
	return &probeBuilder{}
}

func (b *probeBuilder) reset() {
	b.options = probeOptions{}
	b.headers = probeHeaders{}
}

func (b *probeBuilder) SetHeaders(headers []string) error {
	b.headers.protocol = headers[0]
	if len(headers) == 2 {
		var ports []uint16
		for _, pStr := range parser.GroupToSlice(headers[1]) {
			pInt, err := strconv.ParseInt(pStr, 0, 16)
			if err != nil {
				return fmt.Errorf("failed to parse ports: %v", err)
			}
			ports = append(ports, uint16(pInt))
		}
		b.headers.ports = ports
	}
	return nil
}

func (b *probeBuilder) parseFlags(f string) (map[string]string, error) {
	sim := &parser.Symbols{}
	flags, err := sim.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("failed to parse flags: %v", err)
	}
	return flags, nil
}

func (b *probeBuilder) SetOptions(options map[string]string) error {
	var burnt []string

	for key, val := range options {
		if slices.Contains(burnt, key) {
			return fmt.Errorf("found duplicated option: %v", key)
		}
		burnt = append(burnt, key)

		switch key {
		case "sid":
			b.options.sid = val
		case "flags":
			flags, err := b.parseFlags(val)
			if err != nil {
				return fmt.Errorf("failed to parse probe flags: %v", err)
			}
			b.options.flags = flags
		default:
			return fmt.Errorf("probe option not allowed: %s", key)
		}
	}

	// use the protocol as the sid
	if b.options.sid == "" {
		b.options.sid = b.headers.protocol
	}

	return nil
}

func (b *probeBuilder) Build(ac *action) (Probe, error) {
	defer b.reset()
	var probe Probe

	if err := b.SetHeaders(ac.Headers); err != nil {
		return probe, fmt.Errorf("failed to set rule headers '%v': %v", ac.Headers, err)
	}

	if err := b.SetOptions(ac.Options); err != nil {
		return probe, fmt.Errorf("failed to set rule options '%v': %v", ac.Options, err)
	}

	return Probe{
		Sid:      b.options.sid,
		Protocol: b.headers.protocol,
		Ports:    b.headers.ports,
		Flags:    b.options.flags,
	}, nil
}

type action struct {
	Type    string
	Headers []string
	Options map[string]string
}

// TODO: implement this as a Symbol-style parser, closer to a language
// than it is right now. Using regular expressions is not the best
// approach.
type actionParser struct{}

func (p *actionParser) actionType(t string) (string, error) {
	var actionTypes = []string{"scan", "probe", "rule"}
	aType := strings.Trim(t, " ")

	if !slices.Contains(actionTypes, aType) {
		return "", fmt.Errorf("unknown action: %s", t)
	}
	return aType, nil
}

func (p *actionParser) headers(headers string) ([]string, error) {
	const hLen int = 2

	h := make([]string, 0, hLen)
	fields := strings.Fields(headers)
	if len(fields) > hLen {
		return nil, fmt.Errorf("too many headers. Actions must have at most 2 headers: %s", headers)
	}

	for i, field := range fields {
		fields[i] = strings.Trim(field, "\"")
	}

	h = append(h, fields...)
	return h, nil
}

func (p *actionParser) options(ops string) map[string]string {
	matches := OptionsExp.FindAllStringSubmatch(ops, -1)

	var options = make(map[string]string)
	for _, match := range matches {
		key := match[1]
		values := strings.TrimSpace(match[2])
		options[key] = values
	}
	return options
}

func (p *actionParser) Parse(line string) (*action, error) {
	line = strings.TrimSpace(line)
	matches := ActionExp.FindStringSubmatch(line)
	if matches == nil {
		return nil, fmt.Errorf("%w: %s", ErrMalformedAction, line)
	}

	var (
		ac  action
		err error
	)

	if ac.Type, err = p.actionType(matches[1]); err != nil {
		return nil, fmt.Errorf("failed to parse headers: %w", err)
	}

	if ac.Headers, err = p.headers(matches[2]); err != nil {
		return nil, fmt.Errorf("failed to parse headers: %w", err)
	}

	ac.Options = p.options(matches[3])
	return &ac, nil
}

type signatureParser struct {
	action   actionParser
	builders struct {
		rule  *ruleBuilder
		scan  *scanBuilder
		probe *probeBuilder
	}
}

func makeSignatureParser() *signatureParser {
	p := &signatureParser{action: actionParser{}}
	p.builders.rule = newRuleBuilder()
	p.builders.scan = newScanBuilder()
	p.builders.probe = newProbeBuilder()
	return p
}

func (sp *signatureParser) linkRules(sig *Signature, wrappers []ruleWrapper) {
	for i := 0; i < len(sig.Rules); i++ {
		// get the rule and the wrapper
		rule := sig.Rules[i]
		wrapper := wrappers[i]

		// find the rest of the rules this one depends on
		for _, r := range sig.Rules {
			// if the sid is in the deps, add the rule
			if slices.Contains(wrapper.rules, r.Sid) {
				rule.Track = append(rule.Track, r)
			}
		}
	}
}

func (sp *signatureParser) linkScans(sig *Signature, wrappers []scanWrapper) {
	for i := 0; i < len(sig.Scans); i++ {
		scan := sig.Scans[i]
		wrapper := wrappers[i]
		for _, sigRule := range sig.Rules {
			if slices.Contains(wrapper.rules, sigRule.Sid) {
				scan.Rules = append(scan.Rules, sigRule)
			}
		}

		for _, probe := range sig.Probes {
			if slices.Contains(wrapper.probes, probe.Sid) {
				scan.Probes = append(scan.Probes, probe)
			}
		}
	}
}

func (sp *signatureParser) isAction(data []byte) bool {
	// empty
	if len(data) == 0 {
		return false
	}

	// comment
	if bytes.HasPrefix(data, []byte("//")) {
		return false
	}
	return true
}

func (sp *signatureParser) parseActions(reader io.Reader) ([]*action, error) {
	var actions []*action
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		data := scanner.Bytes()
		if !sp.isAction(data) {
			continue
		}

		line := string(data)
		action, err := sp.action.Parse(line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse action '%s': %w", line, err)
		}
		actions = append(actions, action)
	}
	return actions, nil
}

func (sp *signatureParser) Parse(sig *Signature, reader io.Reader) error {
	actions, err := sp.parseActions(reader)
	if err != nil {
		return fmt.Errorf("failed to parse actions: %w", err)
	}

	var (
		rules []ruleWrapper
		scans []scanWrapper
	)

	for _, action := range actions {
		switch action.Type {
		case "rule":
			wrap, err := sp.builders.rule.Wrapper(action)
			if err != nil {
				return fmt.Errorf("failed to build rule: %w", err)
			}
			rules = append(rules, wrap)
			sig.Rules = append(sig.Rules, &wrap.rule)

		case "scan":
			wrap, err := sp.builders.scan.Wrapper(action)
			if err != nil {
				return fmt.Errorf("failed to build scan: %w", err)
			}
			scans = append(scans, wrap)
			sig.Scans = append(sig.Scans, &wrap.scan)

		case "probe":
			probe, err := sp.builders.probe.Build(action)
			if err != nil {
				return fmt.Errorf("failed to build probe: %w", err)
			}
			sig.Probes = append(sig.Probes, &probe)
		}
	}

	sp.linkRules(sig, rules)
	sp.linkScans(sig, scans)
	return nil
}
