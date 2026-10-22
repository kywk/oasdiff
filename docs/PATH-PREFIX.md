# Path Prefix Modification

[English](PATH-PREFIX.md) | [繁體中文](zh-tw/PATH-PREFIX.md)

Sometimes paths prefixes need to be modified, for example, to create a new version:
- /api/v1/...
- /api/v2/...

Oasdiff allows comparison of API specifications with modified prefixes by stripping and/or prepending path prefixes.  
In the example above you could compare the files as follows:
```
oasdiff diff original.yaml new.yaml --strip-prefix-base /api/v1 --prefix-base /api/v2
```
or
```
oasdiff diff original.yaml new.yaml --strip-prefix-base /api/v1 --strip-prefix-revision /api/v2
```
Note that stripping precedes prepending.
