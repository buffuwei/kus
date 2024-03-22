#!/bin/sh
npm run build:prod --prefix frontend
commit=`git rev-parse --short  HEAD`
echo "commit: $commit"
go install -ldflags "-X buffuwei/kus/view.Commit=$commit" 