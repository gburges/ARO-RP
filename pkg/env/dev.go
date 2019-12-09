package env

import (
	"context"
	"fmt"
	"net"
	"os"

	mgmtauthorization "github.com/Azure/azure-sdk-for-go/services/authorization/mgmt/2015-07-01/authorization"
	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"

	"github.com/jim-minter/rp/pkg/api"
	"github.com/jim-minter/rp/pkg/util/azureclient/authorization"
	"github.com/jim-minter/rp/pkg/util/clientauthorizer"
	"github.com/jim-minter/rp/pkg/util/instancemetadata"
	utilpermissions "github.com/jim-minter/rp/pkg/util/permissions"
)

type refreshableAuthorizer struct {
	autorest.Authorizer
	sp *adal.ServicePrincipalToken
}

func (ra *refreshableAuthorizer) Refresh() error {
	return ra.sp.Refresh()
}

type dev struct {
	*prod

	log *logrus.Entry

	permissions     authorization.PermissionsClient
	roleassignments authorization.RoleAssignmentsClient
	applications    graphrbac.ApplicationsClient
}

type Dev interface {
	CreateARMResourceGroupRoleAssignment(context.Context, autorest.Authorizer, *api.OpenShiftCluster) error
}

func newDev(ctx context.Context, log *logrus.Entry) (*dev, error) {
	for _, key := range []string{
		"LOCATION",
		"RESOURCEGROUP",
	} {
		if _, found := os.LookupEnv(key); !found {
			return nil, fmt.Errorf("environment variable %q unset", key)
		}
	}

	armAuthorizer, err := auth.NewClientCredentialsConfig(os.Getenv("AZURE_ARM_CLIENT_ID"), os.Getenv("AZURE_ARM_CLIENT_SECRET"), os.Getenv("AZURE_TENANT_ID")).Authorizer()
	if err != nil {
		return nil, err
	}

	d := &dev{
		log:             log,
		roleassignments: authorization.NewRoleAssignmentsClient(os.Getenv("AZURE_SUBSCRIPTION_ID"), armAuthorizer),
		applications:    graphrbac.NewApplicationsClient(os.Getenv("AZURE_TENANT_ID")),
	}

	d.prod, err = newProd(ctx, log, instancemetadata.NewDev(), clientauthorizer.NewAll())
	if err != nil {
		return nil, err
	}

	d.applications.Authorizer, err = d.FPAuthorizer(ctx, azure.PublicCloud.GraphEndpoint)
	if err != nil {
		return nil, err
	}

	fpAuthorizer, err := d.FPAuthorizer(ctx, azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return nil, err
	}

	d.permissions = authorization.NewPermissionsClient(os.Getenv("AZURE_SUBSCRIPTION_ID"), fpAuthorizer)

	return d, nil
}

func (d *dev) Listen() (net.Listener, error) {
	// in dev mode there is no authentication, so for safety we only listen on
	// localhost
	return net.Listen("tcp", "localhost:8443")
}

func (d *dev) FPAuthorizer(ctx context.Context, resource string) (autorest.Authorizer, error) {
	sp, err := d.fpToken(ctx, resource)
	if err != nil {
		return nil, err
	}

	return &refreshableAuthorizer{autorest.NewBearerAuthorizer(sp), sp}, nil
}

func (d *dev) CreateARMResourceGroupRoleAssignment(ctx context.Context, fpAuthorizer autorest.Authorizer, oc *api.OpenShiftCluster) error {
	d.log.Print("development mode: applying resource group role assignment")

	res, err := d.applications.GetServicePrincipalsIDByAppID(ctx, os.Getenv("AZURE_FP_CLIENT_ID"))
	if err != nil {
		return err
	}

	_, err = d.roleassignments.Create(ctx, "/subscriptions/"+os.Getenv("AZURE_SUBSCRIPTION_ID")+"/resourceGroups/"+oc.Properties.ResourceGroup, uuid.NewV4().String(), mgmtauthorization.RoleAssignmentCreateParameters{
		Properties: &mgmtauthorization.RoleAssignmentProperties{
			RoleDefinitionID: to.StringPtr("/subscriptions/" + os.Getenv("AZURE_SUBSCRIPTION_ID") + "/providers/Microsoft.Authorization/roleDefinitions/c95361b8-cf7c-40a1-ad0a-df9f39a30225"),
			PrincipalID:      res.Value,
		},
	})
	if err != nil {
		var ignore bool
		if err, ok := err.(autorest.DetailedError); ok {
			if err, ok := err.Original.(*azure.RequestError); ok && err.ServiceError != nil && err.ServiceError.Code == "RoleAssignmentExists" {
				ignore = true
			}
		}
		if !ignore {
			return err
		}
	}

	d.log.Print("development mode: refreshing authorizer")
	err = fpAuthorizer.(*refreshableAuthorizer).Refresh()
	if err != nil {
		return err
	}

	// try removing the code below after a bit if we don't hit the error
	permissions, err := d.permissions.ListForResourceGroup(ctx, oc.Properties.ResourceGroup)
	if err != nil {
		return err
	}

	ok, err := utilpermissions.CanDoAction(permissions, "Microsoft.Storage/storageAccounts/write")
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("Microsoft.Storage/storageAccounts/write permission not found")
	}

	return nil
}
