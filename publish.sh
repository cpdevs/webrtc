#!/bin/bash -e
git add .
git commit -m "$1"
git push

VERSION=$(git tag | sort -V | tail -1)

PARTS=(${VERSION//./ })
LENGTH=${#PARTS[@]}
LAST=${PARTS[LENGTH-1]}
INCREMENTED=$(( $LAST + 1 ))
NEWVERSION=""

for s in "${PARTS[@]}"
do
    :
    if [ $s != ${PARTS[LENGTH-1]} ]
    then
        NEWVERSION+=$s.
    else
        NEWVERSION+=$INCREMENTED
    fi
done

git tag -a $NEWVERSION -m "$NEWVERSION"
git push --tags