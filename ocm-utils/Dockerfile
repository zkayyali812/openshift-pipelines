FROM registry.access.redhat.com/ubi9/ubi-minimal:latest

RUN microdnf update -y && microdnf install -y python39 python-pip tar gzip jq wget git make unzip

RUN python3 -m pip install --upgrade pip

RUN python3 -m pip install ansible

# Install oc/kubectl
RUN curl -sLO https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/openshift-client-linux.tar.gz -o openshift-client-linux.tar.gz && \
    tar xzf openshift-client-linux.tar.gz && chmod +x oc && mv oc /usr/local/bin/oc && \
    chmod +x kubectl && mv kubectl /usr/local/bin/kubectl && rm openshift-client-linux.tar.gz

# Install yq
RUN curl -sLO https://github.com/mikefarah/yq/releases/download/v4.16.1/yq_linux_amd64 -o yq_linux_amd64 &&\
    mv yq_linux_amd64 /usr/local/bin/yq && \
    chmod +x /usr/local/bin/yq

# Install go 1.17
RUN curl -sLO https://go.dev/dl/go1.17.6.linux-amd64.tar.gz &&\
    rm -rf /usr/local/go && tar -C /usr/local -xzf go1.17.6.linux-amd64.tar.gz

# Install helm
RUN curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 \
    && chmod 700 get_helm.sh \
    && ./get_helm.sh

# Install aws-cli
RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" && \
    unzip awscliv2.zip && \
    ./aws/install && \
    rm -rf awscliv2.zip && \
    rm -rf aws

# Install rosa-cli
RUN curl -sLO https://mirror.openshift.com/pub/openshift-v4/clients/rosa/latest/rosa-linux.tar.gz -o rosa-linux.tar.gz && \
    tar -xvzf rosa-linux.tar.gz && chmod +x rosa && mv rosa /usr/local/bin/rosa && \
    rm rosa-linux.tar.gz

# Install cm-cli
RUN curl -sLO https://github.com/stolostron/cm-cli/releases/download/v1.0.7/cm_linux_amd64.tar.gz -o cm_linux_amd64.tar.gz && \
    tar -xvzf cm_linux_amd64.tar.gz && chmod +x cm && mv cm /usr/local/bin/cm && \
    rm cm_linux_amd64.tar.gz

# Install ocm cli
RUN mkdir -p ~/bin && \
    curl -Lo ~/bin/ocm https://github.com/openshift-online/ocm-cli/releases/download/v0.1.62/ocm-linux-amd64 && \
    chmod +x ~/bin/ocm && \
    mv ~/bin/ocm /usr/local/bin/ocm

ENV PATH=${PATH}:/usr/local/go/bin
