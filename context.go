package main

import "os"

type gocs_context struct {
	root     string // /usr/aml/my_project/
	rev      string // version control version (what version of gocs
	root_cfg string
}

func (s *gocs_context) init(root string) {
	s.root = "./.gocs"
	s.root_cfg = "./.gocs/GOCSCONF"
	s.rev = "1.0.0"

	// update the .gocs directory
	var err = os.Mkdir(s.root, 0777)
	if err != nil {
		open_file, cfg := os.Open(s.root_cfg)

		if cfg != nil {
			panic(cfg)
		}

		_, _ = open_file.WriteString("++GOCS")
		_, _ = open_file.WriteAt([]byte(s.root), 0)
		_, _ = open_file.WriteAt([]byte(s.rev), int64(len(s.root)))
		_, err_write := open_file.WriteString("--GOCS")

		if err_write != nil {
			panic(err_write)
		}
	}
}
