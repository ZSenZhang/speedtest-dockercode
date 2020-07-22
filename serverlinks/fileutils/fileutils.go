package fileutils

import (
	"io/ioutil"
	"log"
	"os"
	"sort"
)

type ParseFileTsFunc func(string) int64

func SortResultFiles(dir string, tsfunc ParseFileTsFunc, asc int) ([]os.FileInfo, error) {
	resultfiles, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Panic(err)
		return nil, err
	}
	if len(resultfiles) == 0 {
		return resultfiles, nil
	}
	sort.Slice(resultfiles, func(i, j int) bool {
		its := tsfunc(resultfiles[i].Name())
		jts := tsfunc(resultfiles[j].Name())
		if asc == 0 {
			return jts > its
		} else {
			return its > jts
		}
	})
	return resultfiles, nil
}
