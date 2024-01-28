package eaglexport

import (
	"strings"

	"github.com/cockroachdb/errors"
)

func (r *SmartFolderRule) Eval(fileInfo *FileInfo) bool {
	var property string
	if r.Property == "name" {
		// fast path for property name
		property = fileInfo.Name
	} else if r.Property == "type" {
		property = fileInfo.Ext
	} else {
		// slow path, use reflection
		panic(errors.Newf("not supported, property: %v", r.Property))
	}

	switch r.Method {
	case "contain":
		return strings.Contains(property, r.Value)
	case "uncontain":
		return !strings.Contains(property, r.Value)
	case "equal":
		return property == r.Value
	case "unequal":
		return property != r.Value
	default:
		panic(errors.Newf("not supported method: %v", r.Method))
	}

	return false
}

func (c *SmartFolderCondition) Eval(fileInfo *FileInfo) bool {
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
	libraryInfo *LibraryInfo
}

func NewFolderFilter(meta *LibraryInfo) FileDispatcher {
	return FileDispatcher{
		libraryInfo: meta,
	}
}

func (f FileDispatcher) Evaluate(fileInfo *FileInfo) (string, error) {
	for _, folder := range f.libraryInfo.SmartFolders {
		for _, cond := range folder.Conditions {
			if cond.Eval(fileInfo) {
				return folder.Name, nil
			}
		}
	}
	return "", nil
}
