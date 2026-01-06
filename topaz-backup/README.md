# topaz-backup

topaz-backup creates backup of the topaz directory store.


## List plugins

```
topaz-backup
Usage: topaz-backup <command>

topaz backup utility

Commands:
  boltdb    boltdb plugin

Flags:
  -h, --help    Show context-sensitive help.

Run "topaz-backup <command> --help" for more information on a command.

topaz-backup: error: expected "boltdb"
```

NOTES:
 
* Currently only the `boltdb` plugin is exposed

## List input argument of a plugin

```
topaz-backup boltdb

Usage: topaz-backup boltdb --db-file=STRING --backup-dir=STRING

boltdb plugin

Flags:
  -h, --help                 Show context-sensitive help.

      --db-file=STRING       database file path
      --backup-dir=STRING    backup directory path

topaz-backup: error: missing flags: --backup-dir=STRING, --db-file=STRING
```

NOTES:

* In order to list the input arguments for a provider, the plugin name must be provided.

## Execute backup

```
topaz-backup boltdb \
--db-file ~/.local/share/topaz/db/gdrive-v33.db \
--backup-dir ~/.local/share/topaz/backup

/Users/gertd/.local/share/topaz/backup/gdrive-v33-20250731T162842.db
```

NOTES:

* When using the `boltdb` plugin, the topaz-backup command `MUST` be executed on the same machine as were the `topazd` process is running. As the `topazd` process holds the exclusive `read-write` connection to the boltdb database file, the backup process uses a `read-only` connection, to copy the content to a new backup file, and flushed the file state to disk when finished.
