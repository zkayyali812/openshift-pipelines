#!/usr/bin/env python3

import os
import yaml
import json
import requests
from requests.structures import CaseInsensitiveDict

_API_ENDPOINT="https://quay.io/api/v1" # Quay Repository Endpoint

class BearerAuth(requests.auth.AuthBase):
    def __init__(self, token):
        self.token = token
    def __call__(self, r):
        r.headers["Accept"] = "application/json"
        r.headers["Authorization"] = "Bearer " + self.token
        r.payload = {}
        return r

def remove_prefix(text, prefix):
    return text[text.startswith(prefix) and len(prefix):]

def getLatestTagOfBundleByVersion(bundle):
    image=bundle['image']
    image = image.split(':', 1)[0]
    quayRepository = remove_prefix(image, 'quay.io/') # Remove prefix to ensure just the repo remains
    endpoint = _API_ENDPOINT + "/repository/" + quayRepository + "/tag"
    print("Searching for latest tag for bundle {bundle}, version {version}.".format(bundle = quayRepository, version = bundle['version']))
    print(endpoint)
    page = 1
    latestTag = ""
    while latestTag == "":
        params = {"onlyActiveTags": "true", "page": page}
        response = requests.request("GET", endpoint, auth=BearerAuth(os.environ['QUAY_BEARER_TOKEN']), params=params)
        if response.status_code != 200:
            print("Quay API request failed. Check the access token")
            exit
        for tag in response.json()['tags']:
            if tag['name'].startswith("v" + bundle['version']):
                return tag['name']
        if response.json()['has_additional']:
            print("Checking page " + str(page))
            page+=1
        else:
            print("Unable to find tag matching given version.")
            return "UNKNOWN"
        

def main():
    print("Looking for local 'config.yaml' file.")

    configYaml = os.path.join(os.path.dirname(os.path.realpath(__file__)), "config.yaml")
    stream = open(configYaml, 'r')
    bundleConfig = yaml.safe_load(stream)
    for bundle in bundleConfig['bundles']:
        if "autoUpdate" in bundle and bundle["autoUpdate"] == False:
            print("Skipping update of bundle {bundle} version {version} because autoUpdate is set to false.".format(bundle = bundle['operator'], version = bundle['version']))
            continue
        tag = getLatestTagOfBundleByVersion(bundle)
        if tag == "UNKNOWN":
            continue
        latestImage = bundle['image'].split(':', 1)[0] + ":" + tag
        bundle['image'] = latestImage
    with open(configYaml, 'w') as fp:
        yaml.dump(bundleConfig, fp)

if __name__ == "__main__":
    main()
