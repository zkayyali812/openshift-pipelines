# OSD Addon Bundling

This script can help you transition bundles from a imaged bundle format into an OSD addon.


## How to use

To use this script, first ensure the [config.yaml](config.yaml) is properly configured. Bundles can be added or updated directly in this configuration file. Rerunning the script will generate the new addon bundles. Many OSD bundles can be generated at one time. OSD bundles can even be packaged as dependancies using this script as well.

### Format

The [config.yaml](config.yaml) configuration file has the following format - 

```yaml
bundles:
- image: "quay.io/acm-d/acm-operator-bundle:v2.5.0-304" # Bundle Full Image Reference
  version: 2.5.0                                        # Bundle Version
  addonChannel: "beta"                                  # Channel to package the addon in. Must be a standard channel name. Ex. stable, beta, fast
  operator: "advanced-cluster-management"               # Name of the bundle
- image: "quay.io/acm-d/mce-operator-bundle:v2.0.0-150"
  version: 2.0.0
  addonChannel: "beta"
  operator: "multicluster-engine"
  parent: "advanced-cluster-management"                 # If this is set, the bundle will be packaged as a dependancy. To be available in the generated catalogsource of the parent
- image: "quay.io/acm-d/klusterlet-operator-bundle:v2.0.0-110"
  version: 2.0.0
  addonChannel: "beta"
  operator: "klusterlet-operator"
  csvNameOverride: "klusterlet-product"
```

## Running

To run the script do the following - 

```
make bundles
```

## Updating the OSD Addon Bundles

This will generate the new bundles. From here, we must generate a new merge request to the [managed-tenants-bundles](https://gitlab.cee.redhat.com/service/managed-tenants-bundles) repository. 

The following steps can guide you through how to do this -
1. The `managed-tenants-bundles` repository must be forked
2. Copy over the contents of the bundles/* directory to the addons/* directory. See sample command below - 
```bash
$ cp -r bundles/* ~/path/to/managed-tenants-bundles/addons
```
3. Submit merge request. We should also be able to approve these merge requests, provided that our gitlab IDs are present [here](https://gitlab.cee.redhat.com/service/managed-tenants-bundles/-/blob/main/addons/advanced-cluster-management/OWNERS). If onboarding or updating another bundle, ensure your OWNERS file is correct or message #forum-managed-tenants for approval.
4. Once merge request has been submitted, ensure that the associated Pipeline build passes. If it does not, this will need to be corrected before proceeding. Check the logs of the Pipeline for more details. If it passes, an OWNER can /lgtm.
    1. If this is your first time forking the repo, you will need to add their App SRE bot (@devtools-bot) as a Maintainer in your code repo.
5. After the MR to managed-tenants-bundles has been merged, a Pipeline will trigger another build, which will automatically generate a new MR to the [managed-tenants](https://gitlab.cee.redhat.com/service/managed-tenants) repo.
6. If you are an OWNER, approve the generated MR to [managed-tentants](https://gitlab.cee.redhat.com/service/managed-tenants) repository to update the addons once the builds have passed. If the builds fail, see logs for more details.
7. Once the MR to [managed-tenants](https://gitlab.cee.redhat.com/service/managed-tenants) repository has been merged, it can take up to 30 minutes before these changes are visible in Staging or Integration environments.
