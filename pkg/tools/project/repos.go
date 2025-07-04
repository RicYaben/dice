package project

import (
	"github.com/dice"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RepoConfiguration struct {
	Filename string
	Driver   string
}

type projectRepo struct {
	dice.Repository
}

func ProjectRepo(dbfpath string) (*projectRepo, error) {
	conf := dice.DatabaseConfiguration{
		// TODO: Add prometheus plugin to this one
		// db.Use(prometheus.New(prometheus.Config{}))
		Filepath: dbfpath,
		Config: &gorm.Config{
			SkipDefaultTransaction: true,
			PrepareStmt:            true,
		},
		Models: []any{&Project{}, &Experiment{}, &Source{}},
	}

	repo, err := dice.NewRepository(conf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open connection to project database")
	}
	return &projectRepo{repo}, nil
}

func (r *projectRepo) saveProjects(projects []*Project) error {
	db := r.Database()
	q := db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).
		Omit("Project.Experiments", "Experiment.Sources").
		Create(projects)

	if err := q.Error; err != nil {
		return errors.Wrap(err, "failed to create projects")
	}

	if err := db.Save(projects).Error; err != nil {
		return errors.Wrap(err, "failed to save projects")
	}

	return nil
}

func (r *projectRepo) getOrCreateProject(serial string) (bool, *Project, error) {
	proj := Project{
		Serial: serial,
	}

	db := r.Database()
	result := db.FirstOrCreate(&proj, proj)
	if err := result.Error; err != nil {
		return false, nil, err
	}

	return result.RowsAffected > 0, &proj, nil
}

func (r *projectRepo) findStudy(stSerial string) (*Project, error) {
	proj := Project{
		Serial: stSerial,
	}

	db := r.Database()
	result := db.First(&proj, proj)
	if err := result.Error; err != nil {
		return nil, errors.Wrap(err, "failed to find Study")
	}
	return &proj, result.Error
}

func (r *projectRepo) getOrCreateExperiment(stSerial, expSerial string) (bool, *Experiment, error) {
	_, proj, err := r.getOrCreateProject(stSerial)
	if err != nil {
		return false, nil, err
	}

	exp := Experiment{
		ProjectID: proj.ID,
		Serial:    expSerial,
	}

	db := r.Database()
	result := db.FirstOrCreate(&exp, exp)
	if err := result.Error; err != nil {
		return false, nil, err
	}

	return result.RowsAffected > 0, &exp, nil
}

// E.g.: s1234/m123/zmap.tar
func (r *projectRepo) getOrCreateSource(projSerial, expSerial, soSerial string) (bool, *Source, error) {
	_, exp, err := r.getOrCreateExperiment(projSerial, expSerial)
	if err != nil {
		return false, nil, err
	}

	so := Source{
		ExperimentID: exp.ID,
		Serial:       soSerial,
	}

	db := r.Database()
	result := db.FirstOrCreate(&so, so)
	if err := result.Error; err != nil {
		return false, nil, err
	}

	return result.RowsAffected > 0, &so, nil
}
