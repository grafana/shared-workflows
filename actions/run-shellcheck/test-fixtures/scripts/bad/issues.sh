#!/bin/bash

echo $UNQUOTED_VAR
arr=(one two three)
for i in ${arr[@]}; do
  echo $i
done
