package dice

import (
	"fmt"
	"net/rpc"
	"os/exec"

	"github.com/hashicorp/go-plugin"
)

type ScanPlugin interface {
	Scan(h *Host) ([]Record, error)
}

type scanRPC struct {
	client *rpc.Client
}

func (g *scanRPC) Scan(h *Host) ([]Record, error) {
	var records []Record
	if err := g.client.Call("Plugin.Scan", h, &records); err != nil {
		return nil, err
	}
	return records, nil
}

type ScanRPCServer struct {
	Impl ScanPlugin
}

func (s *ScanRPCServer) Scan(host *Host, records *[]Record) error {
	r, err := s.Impl.Scan(host)
	if err != nil {
		return err
	}
	*records = r
	return nil
}

type scanPlugin struct {
	Impl ScanPlugin
}

func (p *scanPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &ScanRPCServer{Impl: p.Impl}, nil
}

func (p *scanPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &scanRPC{client: c}, nil
}

type RulePlugin interface {
	Label(r *Record) (Label, error)
}

type ruleRPC struct {
	client *rpc.Client
}

func (g *ruleRPC) Label(r *Record) (Label, error) {
	var label Label
	if err := g.client.Call("Plugin.Label", r, &label); err != nil {
		return label, err
	}
	return label, nil
}

type ruleRPCServer struct {
	Impl RulePlugin
}

func (s *ruleRPCServer) Scan(r *Record, l *Label) error {
	lab, err := s.Impl.Label(r)
	if err != nil {
		return err
	}
	*l = lab
	return nil
}

type rulePlugin struct {
	Impl RulePlugin
}

func (p *rulePlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &ruleRPCServer{Impl: p.Impl}, nil
}

func (p *rulePlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ruleRPC{client: c}, nil
}

var pluginMap = map[string]plugin.Plugin{
	"scan": &scanPlugin{},
	"rule": &rulePlugin{},
}

type dicePlugin struct {
	raw    any
	client *plugin.Client
}

type pluginFactory struct {
	modulesPath string
	plugins     map[string]plugin.Plugin
	registered  map[Module]*dicePlugin
}

func newPluginFactory(path string) *pluginFactory {
	return &pluginFactory{
		modulesPath: path,
		plugins:     pluginMap,
		registered:  make(map[Module]*dicePlugin),
	}
}

// Makes a new client for some specific plugin
func (f *pluginFactory) loadPlugin(m Module) (*dicePlugin, error) {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion: 1,
		},
		Plugins: pluginMap,
		Cmd:     exec.Command(fmt.Sprintf("%s/%s/%s", f.modulesPath, m.Type, m.Source)),
	})

	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	raw, err := rpcClient.Dispense(m.Name)
	if err != nil {
		return nil, err
	}

	p := &dicePlugin{
		client: client,
		raw:    raw,
	}

	f.registered[m] = p
	return p, nil
}

func (f *pluginFactory) getModule(m Module) *dicePlugin {
	return f.registered[m]
}
