package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"
)

const (
	gocsNotTracked  = 02
	gocsAccessError = 01
	gocsOk          = 00
	gocsFatalError  = -01
)

type gocsContext struct {
	root     string // /usr/aml/my_project/
	rev      string // version control version (what version of gocs
	rootCfg  string // the version control root config path.
	files    []string
	rootFile *os.File
}

const (
	gocsMaxPath = 4096
)

func gocsReadConf(ctx *gocsContext) int {
	offset, err := ctx.rootFile.Seek(0, 2)

	if err != nil {
		return gocsAccessError
	}

	if offset < 1 {
		return gocsFatalError
	}

	ctx.files = make([]string, offset)

	_, err = ctx.rootFile.Seek(0, 0)

	if err != nil {
		return gocsFatalError
	} // why did they deprecate SEEK_SET ???? what the fuck

	// iterate over the newly allocated 'files' array
	// and then populate it
	for fileIndex, _ := range ctx.files {
		var path []byte = make([]byte, gocsMaxPath)
		_, err := ctx.rootFile.Read(path)

		if gocsIsAscii(string(path)) {
			break
		}

		if err != nil {
			return gocsAccessError
		}

		ctx.files[fileIndex] = string(path)
	}

	return gocsOk
}

func (s *gocsContext) init(root string) int {
	s.root = root + ".gocs"
	s.rootCfg = root + ".gocs/conf"
	s.rev = "1.0.0"

	// update the .gocs directory
	_ = os.Mkdir(s.root, 0777)

	var err error
	s.rootFile, err = os.OpenFile(s.rootCfg, os.O_RDWR, 0)

	if err != nil {
		s.rootFile, err = os.Create(s.rootCfg)

		if err != nil {
			return gocsAccessError
		}

		err = os.Mkdir(s.root+"/track/", 0777)

		if err != nil {
			return gocsFatalError
		}
	}

	return gocsReadConf(s)
}

func gocsWriteTracked(fp *os.File, diff []byte) int {
	_, err := fp.Write(diff)

	if err != nil {
		return gocsFatalError
	}

	return gocsOk
}

func gocsUpdateTrack(s *gocsContext, filename string, diff []byte) int {
	var fullPath = s.root + "/track/" + filename

	_, err := os.ReadFile(fullPath)
	var fp *os.File

	if err != nil {
		fp, err = os.Create(fullPath)

		if err != nil {
			// there is a directory first. Create the same.
			var substr string
			buf := bytes.NewBufferString(substr)

			for chIndex, ch := range filename {
				if filename[chIndex] == '/' {
					break
				}

				buf.WriteByte(byte(ch))
			}

			substr = s.root + "/track/" + buf.String()
			err = os.Mkdir(substr, 0777)

			fp, err = os.Create(fullPath)

			if err != nil {
				log.Fatal(err)
				return gocsFatalError
			}
		}

		return gocsWriteTracked(fp, diff)
	} else {
		fp, err = os.OpenFile(fullPath, os.O_WRONLY, 0644)
	}

	return gocsWriteTracked(fp, diff)
}

func gocsAddToTrackingList(tracked []string, trackedName string) {
	tracked = append(tracked, trackedName)
}

func (s *gocsContext) track(filename string) int {
	diff, err := os.ReadFile(filename)

	if err != nil {
		return gocsFatalError
	}

	for i, file := range s.files {
		if strings.Contains(file, filename) {
			s.files = append(s.files[:i], s.files[i+1:]...)
			s.files = append(s.files[:i], "--/++:"+filename+"\n")

			if err != nil {
				return gocsAccessError // well unmount has been called I guess...
			}

			fmt.Printf("--/++ %s\n", filename)

			return gocsUpdateTrack(s, filename, diff)
		}
	}

	_, err = s.rootFile.WriteString("++:" + filename + "\n")

	if err != nil {
		fmt.Println(err)
		return gocsFatalError
	}

	gocsAddToTrackingList(s.files, filename)

	fmt.Printf("++ %s\n", filename)
	gocsUpdateTrack(s, filename, diff)

	return gocsOk
}

func (s *gocsContext) untrack(filename string) int {
	cpy := s.files

	for i, file := range cpy {
		if strings.Contains(file, filename) {
			s.files = append(s.files[:i], s.files[i+1:]...)
			s.files = append(s.files, "--:"+filename)

			return gocsOk
		}
	}

	return gocsNotTracked
}

func gocsIsAscii(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII || !unicode.IsPrint(r) {
			return false
		}
	}

	return true
}

func (s *gocsContext) exit() {
	_, err := s.rootFile.Seek(0, 0)

	if err != nil {
		panic(err)
	}

	for _, file := range s.files {
		_, err := s.rootFile.WriteString(file)

		if err != nil {
			panic(err)
		}
	}
}
