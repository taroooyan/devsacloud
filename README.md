# devsacloud

## Description
This tool can easily operate server on sakura cloud. Create, Boot, Stop, Delete. Also this tool can connet to server that created at sakura cloud with ssh. So user dotn't have to install ssh client. You can develop something apps and tool by use only this tool.
It tool written by golang, So it can run on windows, macOS and linux.

## Install
`go get github.com/taroooyan/devsacloud`

## Config
`cp sample.config.toml config.toml`
Write somethin config

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
  show server info
- -ssh
  connect to crating server


## TODO
- Search server and disk form all zone
- Show server plan and price

## LICENSE
MIT
