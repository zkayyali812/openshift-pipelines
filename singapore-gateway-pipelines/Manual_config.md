# Production OSD Cluster Manual Configuration Steps

1. Navigate to [console.redhat.com](https://console.redhat.com) and sign in
2. Click 'OpenShift' in side nav bar
3. Click 'Create cluster'
4. Click 'Create cluster' again next to 'Red Hat OpenShift Dedicated'
    1. If you don't see 'Create cluster' ensure you are logged into the proper OCM account which is onboarded to the ACM organization
5. Page 1 configuration (Billing Model):
    1. Subscription Type: Annual
    2. Infrastruction Type: Customer Cloud Subscription
6. Page 2 configuration (Cluster settings):
    1. Cloud Provider: AWS
        1. AWS Account ID
        2. AWS Access Key ID
        3. AWS Secret Access Key
    2. Details
        1. Cluster name: sgs-prod
        2. Version: 4.10.9
        3. Region: us-east-1
        4. Availability: Single zone
        5. Monitoring: Enable user workload monitoring is checked
        6. Etcd Encryption: Enable additional Etcd encryption is checked
    3. Machine Pool
        1. Worker node instance type: 4 vCPU 30.5 GiB RAM
        2. Worker node count: 4
7. Page 3 configuration (Networking):
    1. Configuratioon
        1. Cluster Privacy: Public
    2. CIDR Ranges
        1. Leave all as defaults
8. Page 4 configuration (Cluster updates):
    1. Select individual updates
9. Create cluster
10. Wait for it to be Running
11. Add HTPASSWD Auth
    1. Navigate to Clusters 'Access Control' tab
    2. Click 'Identity Providers'
    3. Add HTPasswd identity provider
        1. Username 'kubeadmin'
        2. Password use randomly generated. 
    4. Click add 
    5. Click 'Cluster Roles and Access'
    6. Add 'kubeadmin' as dedicated admin and as cluster-admin
    6. Wait ~10 mins and try to login

