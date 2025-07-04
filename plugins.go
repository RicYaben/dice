package dice

import (
	"fmt"
	"net/rpc"
	"os/exec"

	"github.com/hashicorp/go-plugin"
)

type ScanPlugin interface {
	Scan(h *Host) error
}

type scanRPC struct {
	client *rpc.Client
}

func (g *scanRPC) Scan(h *Host) error {
	return g.client.Call("Plugin.Scan", h, nil)
}

type ScanRPCServer struct {
	Impl ScanPlugin
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

type ClassifierPlugin interface {
	Label(r *Host) error
}

type classifierRPC struct {
	client *rpc.Client
}

func (g *classifierRPC) Label(r *Host) error {
	// maybe return a posible "result", i.e., error + value?
	return g.client.Call("Plugin.Label", r, nil)
}

type classifierRPCServer struct {
	Impl ClassifierPlugin
}

type classifierPlugin struct {
	Impl ClassifierPlugin
}

func (p *classifierPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &classifierRPCServer{Impl: p.Impl}, nil
}

func (p *classifierPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &classifierRPC{client: c}, nil
}

type IdentifierPlugin interface {
	Fingerprint(r *Source) error
}

type identifierRPC struct {
	client *rpc.Client
}

type identifierRPCServer struct {
	Impl IdentifierPlugin
}

func (g *classifierRPC) Fingerprint(s *Source) error {
	return g.client.Call("Plugin.Fingerprint", s, nil)
}

func (s *identifierRPCServer) Scan(so *Source) error {
	return s.Impl.Fingerprint(so)
}

type identifierPlugin struct {
	Impl IdentifierPlugin
}

func (p *identifierPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &identifierRPCServer{Impl: p.Impl}, nil
}

func (p *identifierPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &identifierRPC{client: c}, nil
}

var pluginMap = map[string]plugin.Plugin{
	"scan":       &scanPlugin{},
	"identifier": &identifierPlugin{},
	"classifier": &classifierPlugin{},
}

type dicePlugin struct {
	raw    any
	client *plugin.Client
}

type pluginFactory struct {
	modulesPath string
	plugins     map[string]plugin.Plugin
	registered  map[ModuleModel]*dicePlugin
}

func newPluginFactory(path string) *pluginFactory {
	return &pluginFactory{
		modulesPath: path,
		plugins:     pluginMap,
		registered:  make(map[ModuleModel]*dicePlugin),
	}
}

// Makes a new client for some specific plugin
func (f *pluginFactory) loadPlugin(m ModuleModel) (*dicePlugin, error) {
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

func (f *pluginFactory) getModule(m ModuleModel) *dicePlugin {
	return f.registered[m]
}
