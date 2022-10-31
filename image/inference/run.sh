#!/bin/bash

# set -euo pipefail

gitlab_endpoint=${GITLAB_ENDPOINT}
xihe_user=${XIHE_USER}
xihe_user_token=${XIHE_USER_TOKEN}
project_name=${PROJECT_NAME}

obs_ak=${OBS_AK}
obs_sk=${OBS_SK}
obs_endpoint=${OBS_ENDPOINT}
obs_util=${OBS_UTIL_PATH}
obs_lfs_path=${OBS_BUCKET}/${OBS_LFS_PATH} #$OBS_LFS_PATH has no suffix of /.

inference_dir=$project_name/inference
repo_url=http://${xihe_user}:${xihe_user_token}@${gitlab_endpoint}
app=app.py

# workspace
dir=$(pwd)
work_dir=$dir/{workspace}
test -d $work_dir || mkdir $work_dir
cd $work_dir

# helper
download_model {
    local owner=$1
    local repo=$2
    local file=$repo/$3

    git clone $repo_url/${owner}/${repo}

    # check if lfs
    sha=$(sed -n '/^oid sha256:.\{64\}$/p' "$file")
    if [ -n "$sha" ]; then
        # lfs file

	$obs_util config -i=$obs_ak -k=$obs_sk -e=$obs_endpoint

	# cp file
	# ./obsutil cp obs://bucket-test/test1.txt  /test.txt
	# sha[:2], sha[2:4], sha[4:]
	sha=${sha#"oid sha256:"}
	dst=$work_dir/$inference_dir/$(basename $file)
	$obs_util cp obs://$obs_lfs_path/${sha:0:2}/${sha:2:2}/${sha:4} $dst

	if [ ! -e "$dst" ]; then
            echo "no $dst"
            exit 1
        fi
    else
        # small file
	mv $file $inference_dir
    fi


    # move model file to $inference_dir
}

# download project 
git clone $repo_url/$xihe_user/$project_name

if [ ! -e "$inference_dir/$app" ]; then
    echo "no $app"
    exit 1
fi

f="./$inference_dir/config.json"

if [ -e "$f" -a -s "$f" ]; then
	v=$(python3 $dir/pretrain.py $f)
	if [ $? -ne 0 ]; then
		echo $v
		exit 1
	fi

	if [ -n "$v" ]; then
		owner=$(echo $v | sed -n '1p')
		repo=$(echo $v | sed -n '2p')
		pretain_file=$(echo $v | sed -n '3p')

                download_model $owner $repo $pretain_file
	fi
fi

# run
cd $inference_dir

python3 ./$app
