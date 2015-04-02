Charm Sync
==========

Simple tool that consumes a json file and pulls down upstream charm and dependencies and updates the charm on launchpad.


Installation
============

```
go get github.com/gabriel-samfira/charmsync
```

Usage
=====

Create a ```charmsync.json``` file inside an empty folder with content similar to:

```json
{
    "name": "nova-hyperv",
    "scm": "git",
    "url": "https://github.com/cloudbase/nova-hyperv-charm",
    "revision": "6bf9415ddf8a734b250183839982aeb0f6aabba4",
    "upstream": " lp:~gabriel-samfira/charms/win2012hvr2/nova-hyperv-charm/trunk",
    "dependencies": [
        {
            "url": "https://github.com/cloudbase/juju-powershell-modules.git",
            "resources": [
                "CharmHelpers"
            ],
            "destination": "hooks/Modules",
            "scm": "git",
            "name": "juju-powershell-modules",
            "revision": "1fae7240ef29154b666fdac8c85a8fc47e90a356"
        }
    ]
}
```

Run ```charmsync```

This will create a folder structure like:

```
directory
|-- dependencies
   |-- juju-powershell-modules
|-- nova-hyperv-charm # this is your development branch
|-- staging
   |-- nova-hyperv-charm # this is the branch on launchpad published in the charm store
```
