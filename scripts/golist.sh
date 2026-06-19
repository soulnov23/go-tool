#!/bin/bash

set -x
set -e

go list -deps . | grep -E 'backcalllog|calllog|runlog|common_lib'
