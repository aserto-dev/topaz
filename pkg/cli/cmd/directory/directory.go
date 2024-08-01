package directory

type DirectoryCmd struct {
	Check   CheckCmd   `cmd:"" help:"check permission"`
	Search  SearchCmd  `cmd:"" help:"search relation graph"`
	Get     GetCmd     `cmd:"" help:"get object|relation|manifest"`
	Set     SetCmd     `cmd:"" help:"set object|relation|manifest"`
	Delete  DeleteCmd  `cmd:"" help:"delete object|relation|manifest"`
	List    ListCmd    `cmd:"" help:"list objects|relations"`
	Import  ImportCmd  `cmd:"" help:"import directory data"`
	Export  ExportCmd  `cmd:"" help:"export directory data"`
	Backup  BackupCmd  `cmd:"" help:"backup directory data"`
	Restore RestoreCmd `cmd:"" help:"restore directory data"`
	Stats   StatsCmd   `cmd:"" help:"directory statistics"`
	Test    TestCmd    `cmd:"" help:"execute directory assertions"`
}

type GetCmd struct {
	Object   GetObjectCmd   `cmd:"" help:"get object"`
	Relation GetRelationCmd `cmd:"" help:"get relation"`
	Manifest GetManifestCmd `cmd:"" help:"get manifest"`
}

type SetCmd struct {
	Object   SetObjectCmd   `cmd:"" help:"set object"`
	Relation SetRelationCmd `cmd:"" help:"set relation"`
	Manifest SetManifestCmd `cmd:"" help:"set manifest"`
}

type DeleteCmd struct {
	Object   DeleteObjectCmd   `cmd:"" help:"delete object"`
	Relation DeleteRelationCmd `cmd:"" help:"delete relation"`
	Manifest DeleteManifestCmd `cmd:"" help:"delete manifest"`
}

type ListCmd struct {
	Objects   ListObjectsCmd   `cmd:"" help:"list objects"`
	Relations ListRelationsCmd `cmd:"" help:"list relations"`
}
