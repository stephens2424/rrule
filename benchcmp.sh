#!/bin/bash

NEW=$(mktemp)
OLD=$(mktemp)

go test -run xxx -bench RRule > $NEW
go test -run xxx -bench Teambition | sed 's#Teambition#RRule#' > $OLD

benchcmp $OLD $NEW

rm $OLD $NEW
