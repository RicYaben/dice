package test

import (
	"github.com/dice"
	"github.com/spf13/afero"
)

func testConfig() (*dice.Configuration, error) {
	// I would like to call conf.Temporary() or dice.TempConfig()
	// So when we create databases and stuff, is inside that temporary
	// directory that clears on exit
	fs := afero.NewMemMapFs()

	conf, err := dice.LoadSettings("-", dice.DefaultPaths())
	if err != nil {
		return nil, err
	}
	conf.SetFS(fs)
	return conf, nil
}
