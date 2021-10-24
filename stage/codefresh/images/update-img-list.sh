#!/bin/bash

./get-img-list.sh --repo prod > images-list
git commit -am "update the image list"
git push -u origin CR-7223
