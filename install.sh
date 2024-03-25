#!/bin/sh
# When you need to rebuild frontend, you can uncomment the following line
# npm run build:prod --prefix frontend
tag=`git describe --tags --abbrev=0`
echo "tag: $tag"
go install -ldflags "-X buffuwei/kus/view.Version=$tag" 