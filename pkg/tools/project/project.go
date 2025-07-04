package project

import (
	"fmt"

	"github.com/dice"
	"github.com/pkg/errors"
)

type projectManager struct {
	repo  *projectRepo
	store FileStore
}

func (m *projectManager) New(project, experment, source string) error {
	return nil
}
func (m *projectManager) List(project, experment, source string) error {
	return nil
}
func (m *projectManager) Delete(project, experment, source string) error {
	return nil
}

func (m *projectManager) createProject(serial string) error {
	return m.repo.WithTransaction(func(repo dice.Repository) error {
		r := repo.(*projectRepo)
		created, project, err := r.getOrCreateProject(serial)
		if err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}

		if created {
			return fmt.Errorf("project %s already exists", serial)
		}

		if err := project.Save(m.store); err != nil {
			return fmt.Errorf("failed to save project %s: %w", serial, err)
		}
		return nil
	})
}

func (m *projectManager) addExperiment(serial, ms string) error {
	return m.repo.WithTransaction(func(repo dice.Repository) error {
		r := repo.(*projectRepo)

		created, msr, err := r.getOrCreateExperiment(serial, ms)
		if err != nil {
			return fmt.Errorf("failed to create experiment: %w", err)
		}

		if created {
			return fmt.Errorf("experiment %s already exists", ms)
		}

		if err := msr.Save(m.store); err != nil {
			return fmt.Errorf("failed to save experiment %s: %w", serial, err)
		}
		return nil
	})
}

func (m *projectManager) addSource(st, ms, so string) error {
	return m.repo.WithTransaction(func(repo dice.Repository) error {
		r := repo.(*projectRepo)

		created, src, err := r.getOrCreateSource(st, ms, so)
		if err != nil {
			return errors.Wrap(err, "failed while creating source")
		}

		if created {
			return errors.Errorf("source %s already exists", so)
		}

		if err := src.Save(m.store); err != nil {
			return errors.Wrapf(err, "failed to save source %s", err)
		}
		return nil
	})
}

func newProjectManager(repo *projectRepo, store FileStore) *projectManager {
	return &projectManager{
		repo:  repo,
		store: store,
	}
}
