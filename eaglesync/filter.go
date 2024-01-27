package eaglesync

import (
	"strings"

	"github.com/cockroachdb/errors"
)

func (r *EagleSmartFolderRule) Eval(fileInfo *EagleFileInfo) bool {
	var property string
	if r.Property == "name" {
		// fast path for property name
		property = fileInfo.Name
	} else {
		// slow path, use reflection
		panic(errors.New("not supported"))
	}

	switch r.Method {
	case "contain":
		return strings.Contains(property, r.Value)
	default:
		panic(errors.New("not supported method"))
	}

	return false
}

func (c *EagleSmartFolderCondition) Eval(fileInfo *EagleFileInfo) bool {
	if len(c.Rules) <= 0 {
		panic(errors.New("Smart folder rules is empty"))
	}

	rcs := make([]bool, 0)
	for _, rule := range c.Rules {
		rcs = append(rcs, rule.Eval(fileInfo))
	}
	ans := rcs[0]

	switch c.Match {
	case "OR":
		for _, b := range rcs[1:] {
			ans = ans || b
		}
	case "AND":
		for _, b := range rcs[1:] {
			ans = ans && b
		}
	default:
		panic(errors.Newf("unexpected MATCH: %v", c.Match))
	}
	expected := c.Boolean == "TRUE"
	return expected == ans
}

type FileDispatcher struct {
	libraryInfo *EagleLibraryInfo
}

func NewFolderFilter(meta *EagleLibraryInfo) FileDispatcher {
	return FileDispatcher{
		libraryInfo: meta,
	}
}

func (f FileDispatcher) Evaluate(fileInfo *EagleFileInfo) (string, error) {
	for _, folder := range f.libraryInfo.SmartFolders {
		for _, cond := range folder.Conditions {
			if cond.Eval(fileInfo) {
				return folder.Name, nil
			}
		}
	}
	return "", nil
}
