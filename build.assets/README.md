### Build Assets
This directory is used by Jenkins to build a sample gravity application for CI/CD testing of changes to the provisioner.

To execute manually run:
```
make OPS_URL=<ops_url> OPS_KEY=<ops_key> TAG=<docker_tag>
```

Where
- OPS_URL: URL of ops center to use for the sample application
- OPS_KEY: token to access the ops center
- TAG: is a tag of the provisioning container available at https://quay.io/repository/gravitational/provisioner?tab=tags
