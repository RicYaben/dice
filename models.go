package dice

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"iter"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Signature struct {
	gorm.Model

	Name string `gorm:"uniqueIndex"`

	Scans  []*Scan  `gorm:"foreignKey:SignatureID;constraint:OnDelete:CASCADE"`
	Rules  []*Rule  `gorm:"foreignKey:SignatureID;constraint:OnDelete:CASCADE"`
	Probes []*Probe `gorm:"foreignKey:SignatureID;constraint:OnDelete:CASCADE"`
}

// TODO: add a field to link rules to scans or probes
// This way, I can react not on any new record, but on
// some specific records, like a filter.
type Rule struct {
	gorm.Model
	SignatureID uint

	// Headers
	Mode   string
	Source string

	// Options
	Sid   string
	Role  string
	Track []*Rule `gorm:"many2many:rule_rules"`
}

type Scan struct {
	gorm.Model

	SignatureID uint

	// Headers
	Protocol string
	Module   string

	// Options
	Mapper Mapper   `gorm:"embedded;embeddedPrefix:map_"`
	Rules  []*Rule  `gorm:"many2many:scan_rules"`
	Probes []*Probe `gorm:"many2many:scan_probes"`
}

type Probe struct {
	gorm.Model

	SignatureID uint

	// Headers
	Protocol string
	Ports    ByteArrayUint16 `gorm:"type:blob"`

	// Options
	Sid   string
	Flags ByteArrayMap `gorm:"type:blob"`
}

// ByteArrayUint16 is a custom type for []uint16 stored as bytes
type ByteArrayUint16 []uint16

func (a ByteArrayUint16) Value() (driver.Value, error) {
	return json.Marshal(a) // Serialize to JSON as []byte
}

func (a *ByteArrayUint16) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), a)
}

// ByteArrayMap is a custom type for map[string]string stored as bytes
type ByteArrayMap map[string]string

func (m ByteArrayMap) Value() (driver.Value, error) {
	return json.Marshal(m) // Serialize to JSON as []byte
}

func (m *ByteArrayMap) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), m)
}

type Mapper struct {
	Mode   string
	Source string
}

type Host struct {
	gorm.Model

	Ip      string
	Domain  string
	Records []Record `gorm:"foreignKey:ID"`
}

func (h *Host) GetRecords(tx *gorm.DB, nodeID uint) (iter.Seq2[*Record, error], error) {
	// Labelled records are already marked, there is no need to filter them
	q := tx.Model(&Record{}).
		Where("records.host_id = ?", h.ID).
		Joins("LEFT JOIN marks ON marks.record_id = records.id AND marks.node_id = ?", nodeID).
		Where("marks.id IS NULL").
		Preload("Marks")

	// Descending because we want to iterate them from the
	// newest to the oldest
	rows, err := q.Order("created_at DESC").Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to search for unmarked records")
	}
	defer rows.Close()

	it := func(rows *sql.Rows) iter.Seq2[*Record, error] {
		return func(yield func(*Record, error) bool) {
			for rows.Next() {
				var record Record
				err := rows.Scan(&record)
				if !yield(&record, err) {
					return
				}
			}
		}
	}
	return it(rows), nil
}

// TODO: somehow this info needs to be added
// Fingerprints map[string]any `json:"figerprints"`
// MD5          string         `json:"md5"` // MD5 of the result, not the whole data object.
type Record struct {
	gorm.Model

	HostID uint
	NodeID uint
	Data   []byte  `gorm:"serializer:json"`
	Labels []Label `gorm:"foreignKey:ID"`
	Marks  []Mark  `gorm:"foreignKey:ID"`
}

func (r *Record) Mark(tx *gorm.DB, nodeID uint) error {
	mark := Mark{
		FingerprintID: r.ID,
		NodeID:        nodeID,
	}
	if err := tx.Create(&mark).Error; err != nil {
		return fmt.Errorf("failed to mark record: %w", err)
	}
	return nil
}

type Label struct {
	gorm.Model

	RecordID    uint
	NodeID      uint
	Annotations []string
	Tags        []string
}

type Mark struct {
	gorm.Model

	FingerprintID uint
	NodeID        uint
}

type Node struct {
	ID          uint
	SignatureID uint
	Type        string
	Module      ModuleModel `gorm:"embedded;embeddedPrefix:node_"`

	// To find leaf nodes find by signatureID without any childs with the same
	// signatureID
	Childs []*Node `gorm:"many2many:node_childs;constraint:OnDelete:CASCADE"`
}

type ModuleModel struct {
	gorm.Model

	Name   string
	Type   string
	Source string
}

type SourceType string

const (
	SourceFile  SourceType = "file"
	SourceStdin SourceType = "stdin"
	SourceArgs  SourceType = "args"
)

type SourceModel struct {
	gorm.Model

	// Name of the source
	Name string
	Type SourceType `gorm:"index"`
	// Data format: json,csv,...
	Format   string
	Location string         // for files or stdin descriptions
	Args     datatypes.JSON // for args type only
}

type Fingerprint struct {
	gorm.Model

	HostID   uint
	RecordID uint
	NodeID   uint
	Data     map[string]any
}

// PROJECT
// ---

// A project is a directory that holds scans with an invididual setup
type Project struct {
	gorm.Model

	// Where the project lives
	Home string
	// Name of the project
	Name string
}
