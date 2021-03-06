package netconf

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestIfacesApplier(t *testing.T) {
	tests := []struct {
		input            string
		expectedOutput   string
		configuratorType BareMetalType
	}{
		{
			input:            "testdata/firewall.yaml",
			expectedOutput:   "testdata/networkd/firewall",
			configuratorType: Firewall,
		},
		{
			input:            "testdata/machine.yaml",
			expectedOutput:   "testdata/networkd/machine",
			configuratorType: Machine,
		},
	}

	tmpPath = os.TempDir()
	for _, tc := range tests {
		func() {
			old := systemdNetworkPath
			tempdir, err := ioutil.TempDir(os.TempDir(), "networkd*")
			systemdNetworkPath = tempdir
			if err != nil {
				log.Fatal(err)
			}
			defer func() {
				os.RemoveAll(systemdNetworkPath)
				systemdNetworkPath = old
			}()
			kb := NewKnowledgeBase(tc.input)
			a := NewIfacesApplier(tc.configuratorType, kb)
			a.Apply()
			if equal, s := equalDirs(systemdNetworkPath, tc.expectedOutput); !equal {
				t.Error(s)
			}
		}()
	}
}

func equalDirs(dir1, dir2 string) (bool, string) {
	files1 := list(dir1)
	files2 := list(dir2)
	if !cmp.Equal(files1, files2) {
		return false, fmt.Sprintf("list of files is different: %v", cmp.Diff(files1, files2))
	}

	for _, f := range files1 {
		f1, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", dir1, f))
		if err != nil {
			panic(err)
		}
		f2, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", dir2, f))
		if err != nil {
			panic(err)
		}
		s1 := string(f1)
		s2 := string(f2)
		if !cmp.Equal(s1, s2) {
			return false, fmt.Sprintf("file %s differs: %v", f, cmp.Diff(s1, s2))
		}
	}
	return true, ""
}

func list(dir string) []string {
	f, err := os.Open(dir)
	if err != nil {
		panic(err)
	}
	finfos, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		panic(err)
	}
	files := []string{}
	for _, file := range finfos {
		files = append(files, file.Name())
	}
	sort.Strings(files)
	return files
}
