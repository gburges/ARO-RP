# Copyright (c) Microsoft Corporation.
# Licensed under the Apache License 2.0.

from azext_aro._validators import validate_cidr
from azext_aro._validators import validate_client_id
from azext_aro._validators import validate_client_secret
from azext_aro._validators import validate_cluster_resource_group
from azext_aro._validators import validate_domain
from azext_aro._validators import validate_pull_secret
from azext_aro._validators import validate_subnet
from azext_aro._validators import validate_visibility
from azext_aro._validators import validate_vnet
from azext_aro._validators import validate_vnet_resource_group_name
from azext_aro._validators import validate_worker_count
from azext_aro._validators import validate_worker_vm_disk_size_gb
from azure.cli.core.commands.parameters import name_type
from azure.cli.core.commands.parameters import resource_group_name_type
from azure.cli.core.commands.parameters import tags_type
from azure.cli.core.commands.validators import get_default_location_from_resource_group


def load_arguments(self, _):
    with self.argument_context('aro') as c:
        c.argument('location',
                   validator=get_default_location_from_resource_group)
        c.argument('resource_name',
                   name_type,
                   help='Name of cluster.')
        c.argument('tags',
                   tags_type)

        c.argument('pull_secret',
                   help='Pull secret of cluster.',
                   validator=validate_pull_secret)
        c.argument('domain',
                   help='Domain of cluster.',
                   validator=validate_domain)
        c.argument('cluster_resource_group',
                   help='Resource group of cluster.',
                   validator=validate_cluster_resource_group)

        c.argument('client_id',
                   help='Client ID of cluster service principal.',
                   validator=validate_client_id)
        c.argument('client_secret',
                   help='Client secret of cluster service principal.',
                   validator=validate_client_secret)

        c.argument('pod_cidr',
                   help='CIDR of pod network.',
                   validator=validate_cidr('pod_cidr'))
        c.argument('service_cidr',
                   help='CIDR of service network.',
                   validator=validate_cidr('service_cidr'))

        c.argument('master_vm_size',
                   help='Size of master VMs.')

        c.argument('worker_vm_size',
                   help='Size of worker VMs.')
        c.argument('worker_vm_disk_size_gb',
                   help='Disk size in GB of worker VMs.',
                   validator=validate_worker_vm_disk_size_gb)
        c.argument('worker_count',
                   help='Count of worker VMs.',
                   validator=validate_worker_count)

        c.argument('apiserver_visibility',
                   help='API server visibility.',
                   validator=validate_visibility('apiserver_visibility'))

        c.argument('ingress_visibility',
                   help='Ingress visibility.',
                   validator=validate_visibility('ingress_visibility'))

        c.argument('vnet_resource_group_name',
                   resource_group_name_type,
                   options_list=['--vnet-resource-group'],
                   help='Name of vnet resource group.',
                   validator=validate_vnet_resource_group_name)
        c.argument('vnet',
                   help='Name or ID of vnet.  If name is supplied, `--vnet-resource-group` must be supplied.',
                   validator=validate_vnet)
        c.argument('master_subnet',
                   help='Name or ID of master vnet subnet.  If name is supplied, `--vnet` must be supplied.',
                   validator=validate_subnet('master_subnet'))
        c.argument('worker_subnet',
                   help='Name or ID of worker vnet subnet.  If name is supplied, `--vnet` must be supplied.',
                   validator=validate_subnet('worker_subnet'))
