package database

import (
	"database/sql"
	"fmt"
	"iter"
	"slices"

	"gorm.io/gorm"
)

type Host struct {
	gorm.Model

	Ip      string
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
	ScanID uint
	Data   []byte  `gorm:"serializer:json"`
	Labels []Label `gorm:"foreignKey:ID"`
	Marks  []Mark  `gorm:"foreignKey:ID"`
}

func (r *Record) Mark(tx *gorm.DB, nodeID uint) error {
	mark := Mark{
		RecordID: r.ID,
		NodeID:   nodeID,
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

	RecordID uint
	NodeID   uint
}

type Node struct {
	gorm.Model

	SignatureID uint
	ModuleID    uint
	Module      Module

	// To find leaf nodes find by signatureID without any childs with the same
	// signatureID
	Childs []*Node `gorm:"many2many:node_childs;constraint:OnDelete:CASCADE"`
}

func (n *Node) AddChild(db *gorm.DB, node *Node) error {
	if slices.Contains(n.Childs, node) {
		return fmt.Errorf("node already linked %d", node.ID)
	}
	n.Childs = append(n.Childs, node)

	// TODO: I am not sure I want to save here
	// I think is better if I save after adding all nodes
	// db.Save(n)
	return nil
}

type Module struct {
	gorm.Model

	Name   string
	Type   string
	Source string
}

type Notification struct {
	gorm.Model

	HostID     uint
	NodeID     uint
	ObjectType string
	ObjectID   uint
}
