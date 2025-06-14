package dice

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	withTransaction(fn func(repo Repository) error) error
}

type repository struct {
	db *gorm.DB
}

// do whatever within a separate withTransaction
func (r *repository) withTransaction(fn func(repo Repository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		repo := r
		repo.db = tx
		return fn(repo)
	})
}

type hostRepo struct {
	repository
}

// returns a host by id
func (r *hostRepo) getHost(id uint) (*Host, error) {
	panic("not implemented yet")
}

// returns a record by id
func (r *hostRepo) getRecord(id uint) (*Record, error) {
	panic("not implemented yet")
}

// stores a record
func (r *hostRepo) saveRecord(record Record) error {
	panic("not implemented yet")
}

// stores a label
func (r *hostRepo) saveLabel(label Label) error {
	panic("not implemented yet")
}

// stores a mark
func (r *hostRepo) saveMark(mark Mark) error {
	panic("not implemented yet")
}

type recordRepo struct {
	repository
}

func (r *recordRepo) findUnmarkedRecords(hostID, nodeID uint) ([]*Record, error) {
	panic("not implemented yet")
}

type eventRepo struct {
	repository
}

func (r *eventRepo) addEvent(event hostEvent) error {
	panic("not implemented yet")
}

type signatureRepo struct {
	repository
}

func (r *signatureRepo) getSignature(name string) (*Signature, error) {
	panic("not implemented yet")
}

func (r *signatureRepo) findSignatures(names []string) ([]string, error) {
	var found []string
	q := r.db.Model(&Signature{}).Where("name IN ?", names).Pluck("name", &found)
	if err := q.Error; err != nil {
		return nil, fmt.Errorf("failed to find names: %w", err)
	}
	return found, nil
}

func (r *signatureRepo) saveSignatures(signatures []*Signature) error {
	if len(signatures) == 0 {
		// no signatures to store
		return nil
	}

	q := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		UpdateAll: true,
	}).
		Omit("Scans.Probes", "Scans.Rules", "Rules.Track").
		Create(signatures)

	if err := q.Error; err != nil {
		return fmt.Errorf("failed to create signatures: %w", err)
	}

	if err := r.db.Save(signatures).Error; err != nil {
		return fmt.Errorf("failed to save signatures: %w", err)
	}
	return nil
}

func (r *signatureRepo) deleteSignatures() error {
	q := r.db.Session(&gorm.Session{AllowGlobalUpdate: true})
	if err := q.Unscoped().Select(clause.Associations).Delete(&Signature{}).Error; err != nil {
		return fmt.Errorf("failed to delete signatures with associations: %w", err)
	}
	return nil
}
