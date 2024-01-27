package eaglexport

type Mtime = map[string]int64

type FileInfo struct {
	ID               string    `json:"id,omitempty"`
	Name             string    `json:"name,omitempty"`
	Size             int       `json:"size,omitempty"`
	Btime            int64     `json:"btime,omitempty"`
	Mtime            int64     `json:"mtime,omitempty"`
	Ext              string    `json:"ext,omitempty"`
	Tags             []any     `json:"tags,omitempty"`
	Folders          []any     `json:"folders,omitempty"`
	IsDeleted        bool      `json:"isDeleted,omitempty"`
	URL              string    `json:"url,omitempty"`
	Annotation       string    `json:"annotation,omitempty"`
	ModificationTime int64     `json:"modificationTime,omitempty"`
	Height           int       `json:"height,omitempty"`
	Width            int       `json:"width,omitempty"`
	Palettes         []Palette `json:"palettes,omitempty"`
	DeletedTime      int64     `json:"deletedTime,omitempty"`
	LastModified     int64     `json:"lastModified,omitempty"`
}

type Palette struct {
	Color []int   `json:"color"`
	Ratio float32 `json:"ratio"`
}

type SmartFolder struct {
	ID               string                 `json:"id,omitempty"`
	Name             string                 `json:"name,omitempty"`
	Description      string                 `json:"description,omitempty"`
	ModificationTime int64                  `json:"modificationTime,omitempty"`
	Conditions       []SmartFolderCondition `json:"conditions,omitempty"`
	Children         []any                  `json:"children,omitempty"`
}

type SmartFolderCondition struct {
	Rules   []SmartFolderRule `json:"rules"`
	Match   string            `json:"match"`
	Boolean string            `json:"boolean"`
}

type SmartFolderRule struct {
	Property string `json:"property"`
	Method   string `json:"method"`
	Value    string `json:"value"`
}

type LibraryInfo struct {
	Folders            []any         `json:"folders,omitempty"`
	SmartFolders       []SmartFolder `json:"smartFolders,omitempty"`
	QuickAccess        []any         `json:"quickAccess,omitempty"`
	TagsGroups         []any         `json:"tagsGroups,omitempty"`
	ModificationTime   int64         `json:"modificationTime,omitempty"`
	ApplicationVersion string        `json:"applicationVersion,omitempty"`
}
