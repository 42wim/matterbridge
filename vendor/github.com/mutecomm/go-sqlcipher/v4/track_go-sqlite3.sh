#!/bin/sh -e

if [ $# -ne 1 ]
then
  echo "Usage: $0 go-sqlite3_dir" >&2
  echo "Copy tracked source files from go-sqlite3 to current directory." >&2
  exit 1
fi

ltd=$1

# copy C files
cp -f $ltd/sqlite3_opt_unlock_notify.c .

# copy Go files
cp -f $ltd/*.go .
rm -rf _example
cp -r $ltd/_example .
rm -rf upgrade
cp -r $ltd/upgrade .

echo "make sure to adjust sqlite3.go with sqlcipher pragmas!"
echo "make sure to adjust import paths in _example directory!"
