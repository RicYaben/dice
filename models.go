package dice

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Signature struct {
	gorm.Model

	// Name of the signature
	Name string `gorm:"uniqueIndex"`
	// Component of signature. This defines the modules it contains
	Component string
	// Linked nodes in the signature
	Nodes []*Node `gorm:"foreignKey:SignatureID;constraint:OnDelete:CASCADE"`
}

type NodeType string

const (
	MODULE_NODE    = "module"
	SIGNATURE_NODE = "signature"
)

// Nodes are the objects linked to signatures.
// Not to confuse with Modules, nodes are only wrappers
// around Modules to allow for links.
// Their ID is different from their object ID.
type Node struct {
	gorm.Model

	// Signature this node belongs to
	SignatureID uint

	// Type of node. It can be another signature or a module
	Type NodeType
	// ID of the signature or module
	ObjectID uint
	//Module Module `gorm:"embedded;embeddedPrefix:node_"`

	// To find leaf nodes find by signatureID without any childs with the same
	// signatureID
	Children []*Node `gorm:"many2many:node_children;constraint:OnDelete:CASCADE"`

	// Name used in the signature. Not useful in the db.
	name string
}

type ModuleType string

const (
	CLASSIFIER_MODULE ModuleType = "classifier"
	IDENTIFIER_MODULE ModuleType = "identifier"
	SCANNER_MODULE    ModuleType = "scanner"
)

// A module references a plugin stored somewhere.
// The type of module defines which component will hold it
// and how it will receive updates
type Module struct {
	gorm.Model

	// Name of the module
	Name string
	// Path to module
	Location string
	// Hash of the plugin
	Hash string
	// Tags
	Tags       []string
	Properties datatypes.JSON
}

// A target-ish. It holds fingerprints and labels
// related to a host during a measurement.
type Host struct {
	gorm.Model

	Ip           string
	Domain       string
	Hooks        []*Hook        `gorm:"foreignKey:ID"`
	Fingerprints []*Fingerprint `gorm:"foreignKey:ID"`
	Labels       []*Label       `gorm:"foreignKey:ID"`
}

// Hook is a reference to a module.
// Modules always hook the objects they see.
// When they are done with the object, the hook is marked as done
// When the object is updated, the hooked modules (not done)
// are notified.
type Hook struct {
	gorm.Model

	// Hooked object
	ObjectID uint
	// Hooked module
	NodeID uint
	// Whether the module is done with the object
	Done bool
}

type Fingerprint struct {
	gorm.Model

	HostID uint
	// Which module made it
	ModuleID uint
	// The fingerprint
	Data datatypes.JSON
	// Hash value of the fingerprint's data

	Hash     string
	Service  string
	Protocol string
	Port     uint16
}

// A scan is a type of command that scanners can interpret
// to lunch a type of scan on a target
// E.g., module: zmap, args: "port: 22 protocol:ssh probe:synack"
type Scan struct {
	gorm.Model

	ModuleID uint
	// List of targets
	Targets []string
	// Argument to pass to the scanner
	Args datatypes.JSON
}

// A classification label assigned to a host
type Label struct {
	gorm.Model

	// ID of the host where this label is placed
	HostID uint
	// short name of the label, e.g., broken-access-control
	// This should be unique in the module
	ShortName string
	// Long name of the label. e.g.,
	// "MQTT broken access control - failed authentication"
	LongName string
	// Description of the label, e.g.,
	// why and when a host is assigned this label
	Description string
	// Mitigation advice
	Mitigation string

	//ModuleID uint
}

type SourceType string

const (
	SourceFile  SourceType = "file"
	SourceStdin SourceType = "stdin"
	SourceArgs  SourceType = "args"
)

// A source points to records that need pre-processing
// This is the type of object identifiers consume to create
// fingerprints and hosts
type Source struct {
	gorm.Model

	// Name of the source
	Name string
	Type SourceType `gorm:"index"`
	// Data format: json,csv,...
	Format   string
	Location string         // for files or stdin descriptions
	Args     datatypes.JSON // for args type only
}

// PROJECT
// ---

// A project is a directory that holds scans with an invididual setup
type Project struct {
	gorm.Model

	// Where the project lives
	Path string
	// Name of the project
	Name string

	Settings datatypes.JSON

	Studies []*Study
}

type Study struct {
	gorm.Model
	Name string
	Path string
}

type EventType uint8

const (
	SOURCE_EVENT EventType = iota
	LABEL_EVENT
	FINGERPRINT_EVENT
	HOST_EVENT
	SCAN_EVENT
)

type Event struct {
	gorm.Model

	// Origin
	NodeID uint
	// The type of event
	Type EventType
	// ID of the object holding the event
	ObjectID uint

	// For direct delivery. Signature names
	// TODO: this should be tags! tags are set as properties in
	// signatures
	Targets []string
}
