#!/usr/bin/env bash

export dockerImgTag=yugioh_card_editor
export dockerCtnName=yugioh_card_editor


docker build --tag=$dockerImgTag .
export isBuiltWell=$?
if [[ $isBuiltWell -eq 0 ]]; then
    echo "built image $dockerImgTag with cache"
else
    echo "error when build image, please try to build --no-cache"
    # docker build --tag=$dockerImgTag --no-cache .;
    # export isBuiltWell=$?
fi
if [[ $isBuiltWell -eq 0 ]]; then
    echo "successfully built $dockerImgTag"
else
    echo "fail to build $dockerImgTag"
    exit 1
fi


# clean up old container
echo "stop and remove old container $dockerCtnName"
docker stop $dockerCtnName 2>/dev/null;
docker rm $dockerCtnName 2>/dev/null;


echo "preparing envs"
export envFile=configs/env.sh  # bash script has "export" commands
export dockerRunEnv=configs/.env  # generated from the envFile with the following "sed" for "docker run"
bash -x $envFile 2>$dockerRunEnv
sed -i 's/+ //' ${dockerRunEnv}; sed -i '/^export /d' ${dockerRunEnv}; sed -i "s/'//g" ${dockerRunEnv}

docker run -dit --name=$dockerCtnName --restart always \
    -p 20808:20808 \
    --env-file ${dockerRunEnv} \
    ${dockerImgTag}

echo "done run $dockerCtnName container"
