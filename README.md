# dnalc gearman workers

These will need to be set:

```
export DB_DATABASE=...
export DB_HOST=...
export DB_USER=...
export DB_PASS=...
export GEARMAN_SERVERS=server1,server2 # comma separated
```

## svnupdater

This worker updates svn on the live site. This will be running as root.
The client can request an update like this:

```
my $client = Gearman::Client->new;
$client->job_servers($GEARMAN_SERVERS);
my $res = $client->do_task("SVNUpdate", $site);
```

where $site can be one of the following:
  - dnalc
  - dnabarcoding101
  - learnaboutsma
  - dnaftb
  - summercamps

## cmssynchronizer

This worker performs two things for a given atom id:

- based on data in the CMS (atom_downloads), copies the files from the
live site (or updates them)
- fixes the permissions at the file level:
  * `chmod -Rh a+r` for the files
  * `chmod -Rh a+rx` for the directories
  * `chcon -R -t httpd_sys_rw_content_t` for the atom's directory

This worker will be running on the _content_ server as the _biomedia_ user

The client will invoking one of the two methods/functions like this:
  * `FixAtomPems(atom_id)`
  * `SynchAtomFiles(atom_id)`

