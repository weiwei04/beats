#!/bin/bash

# usage: ./mkpkg.sh <version>

if [[ ${1:-} ]]
then
    echo "version v$1"
else
    echo "Error: Please add version number."
    exit -1
fi

GOPATH_BAK=$GOPATH
#source $QBOXROOT/base/env.sh
cd ../
export GOPATH=$GOPATH:`pwd`/vendor/github.com/qbox/pandora-sdk
cd -

go build

COMMIT=`git rev-parse HEAD`
REVISION=${COMMIT:0:6}
BASEDIR=v$1.$REVISION
cp -r product $BASEDIR
mv filebeat $BASEDIR/bin/logbeat
tar czf logbeat.$BASEDIR.tar.gz $BASEDIR
rm -r $BASEDIR
md5sum logbeat.$BASEDIR.tar.gz > logbeat.$BASEDIR.md5
export GOPATH=$GOPAT_BAK
