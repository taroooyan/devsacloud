# devsacloud

## Install

## Config
`cp sample.config.toml config.toml`
```
token        = "TOKEN"
secret       = "SECRET"
zone         = "tk1a" # tk1a, is1a, is1b
# if name is directory name, don't have to set name
name         = "libsacloud demo"
description  = "libsacloud demo description"
tag          = "libsacloud-test"
cpu          = 1
mem          = 2
# if name is directory name, don't have to set name
hostName     = "libsacloud-test"
password     = "C8#mf92mp!*s"
sshPublicKey = "ssh-rsa AAAA..."
```

## Usege
Usage of ./devsacloud:
- -boot  
  boot server
- -create  
  create new server
- -delete  
  delete server
- -stop  
  stop server
- -show
- show server info

## TODO
- Search server and disk form all zone
- Show server plan and price
- add ssh option to access server with ssh

## LICENSE
MIT
