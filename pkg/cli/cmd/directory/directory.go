package directory

type FormatVersion int

const (
	V2 FormatVersion = 2
	V3 FormatVersion = 3
)

type DirectoryCmd struct {
	Check  CheckCmd  `cmd:"" help:"check"`
	Search SearchCmd `cmd:"" name:"search" help:"search relation graph"`

	GetObject    GetObjectCmd    `cmd:"" help:"get object"`
	SetObject    SetObjectCmd    `cmd:"" help:"set object"`
	DeleteObject DeleteObjectCmd `cmd:"" help:"delete object"`
	ListObjects  ListObjectsCmd  `cmd:"" help:"list objects"`

	GetRelation    GetRelationCmd    `cmd:"" help:"get relation"`
	SetRelation    SetRelationCmd    `cmd:"" help:"set relation"`
	DeleteRelation DeleteRelationCmd `cmd:"" help:"delete relation"`
	ListRelations  ListRelationsCmd  `cmd:"" help:"list relations"`

	Import  ImportCmd  `cmd:"" help:"import directory data"`
	Export  ExportCmd  `cmd:"" help:"export directory data"`
	Backup  BackupCmd  `cmd:"" help:"backup directory data"`
	Restore RestoreCmd `cmd:"" help:"restore directory data"`
	Test    TestCmd    `cmd:"" help:"execute directory assertions"`
}
