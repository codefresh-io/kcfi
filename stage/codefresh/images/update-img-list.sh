#!/bin/bash

-bash: /kcfi/stage/codefresg/images/get-img-list.sh --repo prod > images-list
git config --global user.name shirtabachii
git commit -am "update the image list"
git push -u origin CR-7223
