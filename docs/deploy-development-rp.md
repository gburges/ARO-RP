# Deploy development RP

## Prerequisites

1. Install [Go 1.13](https://golang.org/dl) or later, if you haven't already.

1. Install a supported version of [Python](https://www.python.org/downloads), if
   you don't have one installed already.  The `az` client supports Python 3.6+.
   A recent Python 3.x version is recommended.

1. Fedora users: install the `gpgme-devel` and `libassuan-devel` packages.

   OSX users: please follow [Prepare your development environment using
   OSX](./prepare-your-development-environment-using-osx.md).

1. Install the [Azure
   CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli), if you
   haven't already.

1. Install [OpenVPN](https://openvpn.net/community-downloads), if you haven't
   already.

1. Log in to Azure:

   ```bash
   az login
   ```

1. Git clone this repository to your local machine:

   ```bash
   go get -u github.com/Azure/ARO-RP/...
   cd ${GOPATH:-$HOME/go}/src/github.com/Azure/ARO-RP
   ```

1. If you don't have access to a shared development environment and secrets,
   follow [prepare a shared RP development
   environment](prepare-a-shared-rp-development-environment.md).

1. Set SECRET_SA_ACCOUNT_NAME to the name of the storage account containing your
   shared development environment secrets and save them in `secrets`:

   ```bash
   SECRET_SA_ACCOUNT_NAME=rharosecrets make secrets
   ```

1. Copy, edit (if necessary) and source your environment file.  The required
   environment variable configuration is documented immediately below:

   ```bash
   cp env.example env
   vi env
   . ./env
   ```

   * LOCATION: Location of the shared RP development environment (default:
     `eastus`).

1. Create your own RP database:

   ```bash
   az group deployment create \
     -g "$RESOURCEGROUP" \
     -n "databases-development-$USER" \
     --template-file deploy/databases-development.json \
     --parameters \
       "databaseAccountName=$COSMOSDB_ACCOUNT" \
       "databaseName=$DATABASE_NAME" \
     >/dev/null
   ```


## Run the RP and create a cluster

1. Source your environment file.

   ```bash
   . ./env
   ```

1. Run the RP

   ```bash
   go run ./cmd/aro rp
   ```

1. Before creating a cluster, it is necessary to fake up the step of registering
   the development resource provider to the subscription:

   ```bash
   curl -k -X PUT \
     -H 'Content-Type: application/json' \
     -d '{"state": "Registered", "properties": {"tenantId": "'"$AZURE_TENANT_ID"'"}}' \
     "https://localhost:8443/subscriptions/$AZURE_SUBSCRIPTION_ID?api-version=2.0"
   ```

1. To create a cluster, follow the instructions in [using `az
   aro`](using-az-aro.md).  Note that as long as the RP_MODE environment
   variable is set to development, the `az aro` client will connect to your
   local RP.

1. The following additional RP endpoints are available but not exposed via `az
   aro`:

   * Delete a subscription, cascading deletion to all its clusters:

     ```bash
     curl -k -X PUT \
       -H 'Content-Type: application/json' \
       -d '{"state": "Deleted", "properties": {"tenantId": "'"$AZURE_TENANT_ID"'"}}' \
       "https://localhost:8443/subscriptions/$AZURE_SUBSCRIPTION_ID?api-version=2.0"
     ```

   * List operations:

     ```bash
     curl -k \
       "https://localhost:8443/providers/Microsoft.RedHatOpenShift/operations?api-version=2019-12-31-preview"
     ```


## Debugging

* SSH to the bootstrap node:

  ```bash
  sudo openvpn secrets/vpn-$LOCATION.ovpn &
  hack/ssh-bootstrap.sh "/subscriptions/$AZURE_SUBSCRIPTION_ID/resourceGroups/$RESOURCEGROUP/providers/Microsoft.RedHatOpenShift/openShiftClusters/$CLUSTER"
  ```

* Get an admin kubeconfig:

  ```bash
  make admin.kubeconfig
  export KUBECONFIG=admin.kubeconfig
  ```

* "SSH" to a cluster node:

  * First, get the admin kubeconfig and `export KUBECONFIG` as detailed above.

  ```bash
  hack/ssh.sh [aro-master-{0,1,2}]
  ```

### Metrics

To run fake metrics socket:
```bash
go run ./hack/monitor
```
