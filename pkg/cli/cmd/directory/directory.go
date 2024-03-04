package directory

type FormatVersion int

const (
	V2 FormatVersion = 2
	V3 FormatVersion = 3
)

type DirectoryCmd struct {
	Import         ImportCmd         `cmd:"" help:"import data into directory"`
	Export         ExportCmd         `cmd:"" help:"export directory data"`
	GetObject      GetObjectCmd      `cmd:"" help:"get object" group:"directory"`
	SetObject      SetObjectCmd      `cmd:"" help:"set object" group:"directory"`
	DeleteObject   DeleteObjectCmd   `cmd:"" help:"delete object" group:"directory"`
	ListObjects    ListObjectsCmd    `cmd:"" help:"list objects" group:"directory"`
	GetRelation    GetRelationCmd    `cmd:"" help:"get relation" group:"directory"`
	SetRelation    SetRelationCmd    `cmd:"" help:"set relation" group:"directory"`
	DeleteRelation DeleteRelationCmd `cmd:"" help:"delete relation" group:"directory"`
	ListRelations  ListRelationsCmd  `cmd:"" help:"list relations" group:"directory"`
	Check          CheckCmd          `cmd:"" help:"check" group:"directory"`
	GetGraph       GetGraphCmd       `cmd:"" help:"get relation graph" group:"directory"`
}
