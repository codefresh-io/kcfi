#!/bin/bash

./get-img-list.sh --repo prod > images-list
echo $date >> images-list

cat images-list
