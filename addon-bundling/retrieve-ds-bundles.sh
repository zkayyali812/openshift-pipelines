#!/bin/bash
set -e

extractBundlesFromImage () {
    _image=$1
    _resultsDir=$2

    echo "== Extracting bundles from image ${_image} =="
    ${_DOCKER_OR_PODMAN} pull "$image"
    ${_DOCKER_OR_PODMAN} save "$image" --output temp.tar
    mkdir -p temp
    tar -xf temp.tar -C temp/

    _bundleURI=(${_image//:/ })
    _bundleURL=${_bundleURI[0]}
    _bundleTag=${_bundleURI[1]}

    _bundleDir=$(cat temp/repositories | jq -r ".\"${_bundleURL}\".\"${_bundleTag}\"")
    
    mkdir -p temp/${_bundleDir}/extractedLayer
    _extractedBundle=temp/${_bundleDir}/extractedLayer
    tar -xf temp/${_bundleDir}/layer.tar -C ${_extractedBundle} 

    echo "Image contents extracted. Copying bundle contents to results directory"
    
    mkdir -p ${_resultsDir}/manifests
    mkdir -p ${_resultsDir}/metadata
        
    cp -r ${_extractedBundle}/manifests/*.yaml ${_resultsDir}/manifests/
    cp -r ${_extractedBundle}/metadata/*.yaml ${_resultsDir}/metadata/

    echo "Results successfully copied. Cleaning up"
    rm -rf temp
    rm temp.tar
}

overrideMetadata () {
    _resultsDir=$1
    _channel=$2

    _CHANNEL=$_channel yq eval -i '.annotations."operators.operatorframework.io.bundle.channels.v1" = env(_CHANNEL)' ${_resultsDir}/metadata/annotations.yaml
    _CHANNEL=$_channel yq eval -i '.annotations."operators.operatorframework.io.bundle.channel.default.v1" = env(_CHANNEL)' ${_resultsDir}/metadata/annotations.yaml
}

fixCSVs () {
    _resultsDir=$1
    _csvName=$2
    yq eval -i 'del(.spec.customresourcedefinitions.owned.[].group)' ${_resultsDir}/manifests/${_csvName}
}


_DOCKER_OR_PODMAN=podman
if ! command -v ${_DOCKER_OR_PODMAN} &> /dev/null
then
    _DOCKER_OR_PODMAN=docker
fi

rm -rf bundles/*

declare -a bundles=($(yq e -o=j -I=0 '.bundles[]' config.yaml ))
for bundle in "${bundles[@]}"; do

    image=$(echo "$bundle" | yq e '.image' -)
    version=$(echo "$bundle" | yq e '.version' -)
    addonChannel=$(echo "$bundle" | yq e '.addonChannel' -)
    operator=$(echo "$bundle" | yq e '.operator' -)
    parent=$(echo "$bundle" | yq e '.parent' -)
    csvNameOverride=$(echo "$bundle" | yq e '.csvNameOverride' -)

    if [[ "$parent" != "null" ]]; then
        resultsDir="bundles/${parent}/${operator}/${version}"
    else
        resultsDir="bundles/${operator}/main/${version}"
    fi

    if [[ "$csvNameOverride" != "null" ]]; then
        csvName=${csvNameOverride}.v${version}.clusterserviceversion.yaml
    else
        csvName=${operator}.v${version}.clusterserviceversion.yaml
    fi

    extractBundlesFromImage $image $resultsDir
    overrideMetadata $resultsDir $addonChannel
    fixCSVs $resultsDir $csvName

    echo ""
done
