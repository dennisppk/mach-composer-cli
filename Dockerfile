ARG PYTHON_VERSION="3.8.5"
FROM python:${PYTHON_VERSION}-alpine

ENV AZURE_CLI_VERSION=2.5.1
ENV TERRAFORM_VERSION=0.13.4
ENV TERRAFORM_EXTERNAL_VERSION=1.2.0
ENV TERRAFORM_AZURE_VERSION=2.29.0
ENV TERRAFORM_AWS_VERSION=3.8.0
ENV TERRAFORM_NULL_VERSION=2.1.2
ENV TERRAFORM_COMMERCETOOLS_VERSION=0.23.0
ENV TERRAFORM_SENTRY_VERSION=0.6.0

RUN apk add --no-cache --virtual .build-deps g++ libffi-dev openssl-dev wget unzip jq make curl \
    && apk add bash ca-certificates git libc6-compat openssl openssh-client
    
# Install Azure CLI
RUN pip --no-cache-dir install azure-cli==${AZURE_CLI_VERSION}

# Pre-install Terreform plugins
ENV TERRAFORM_PLUGINS_PATH=/root/.terraform.d/plugins/linux_amd64
RUN mkdir -p ${TERRAFORM_PLUGINS_PATH}

# Install terraform
RUN cd /tmp && \
    wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
    unzip -n terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/bin && \
    rm -rf /tmp/* && \
    rm -rf /var/cache/apk/* && \
    rm -rf /var/tmp/*

# Install null provider
RUN cd /tmp && \
    wget https://releases.hashicorp.com/terraform-provider-null/${TERRAFORM_NULL_VERSION}/terraform-provider-null_${TERRAFORM_NULL_VERSION}_linux_amd64.zip && \
    unzip -n terraform-provider-null_${TERRAFORM_NULL_VERSION}_linux_amd64.zip -d ${TERRAFORM_PLUGINS_PATH} && \
    rm -rf /tmp/* && \
    rm -rf /var/cache/apk/* && \
    rm -rf /var/tmp/*

# Install external provider
RUN cd /tmp && \
    wget https://releases.hashicorp.com/terraform-provider-external/${TERRAFORM_EXTERNAL_VERSION}/terraform-provider-external_${TERRAFORM_EXTERNAL_VERSION}_linux_amd64.zip && \
    unzip -n terraform-provider-external_${TERRAFORM_EXTERNAL_VERSION}_linux_amd64.zip -d ${TERRAFORM_PLUGINS_PATH} && \
    rm -rf /tmp/* && \
    rm -rf /var/cache/apk/* && \
    rm -rf /var/tmp/*

# Install aws provider
RUN cd /tmp && \
    wget https://releases.hashicorp.com/terraform-provider-aws/${TERRAFORM_AWS_VERSION}/terraform-provider-aws_${TERRAFORM_AWS_VERSION}_linux_amd64.zip && \
    unzip -n terraform-provider-aws_${TERRAFORM_AWS_VERSION}_linux_amd64.zip -d ${TERRAFORM_PLUGINS_PATH} && \
    rm -rf /tmp/* && \
    rm -rf /var/cache/apk/* && \
    rm -rf /var/tmp/*

# Install azure provider
RUN cd /tmp && \
    wget https://releases.hashicorp.com/terraform-provider-azurerm/${TERRAFORM_AZURE_VERSION}/terraform-provider-azurerm_${TERRAFORM_AZURE_VERSION}_linux_amd64.zip && \
    unzip -n terraform-provider-azurerm_${TERRAFORM_AZURE_VERSION}_linux_amd64.zip -d ${TERRAFORM_PLUGINS_PATH} && \
    rm -rf /tmp/* && \
    rm -rf /var/cache/apk/* && \
    rm -rf /var/tmp/*

# Install commercetools provider
RUN cd /tmp && \
    wget https://github.com/labd/terraform-provider-commercetools/releases/download/v${TERRAFORM_COMMERCETOOLS_VERSION}/terraform-provider-commercetools_${TERRAFORM_COMMERCETOOLS_VERSION}_linux_amd64.zip && \
    unzip -n terraform-provider-commercetools_${TERRAFORM_COMMERCETOOLS_VERSION}_linux_amd64.zip -d ${TERRAFORM_PLUGINS_PATH} && \
    rm -rf /tmp/* && \
    rm -rf /var/cache/apk/* && \
    rm -rf /var/tmp/*

# Install sentry provider
RUN cd /tmp && \
    wget https://github.com/jianyuan/terraform-provider-sentry/releases/download/v${TERRAFORM_SENTRY_VERSION}/terraform-provider-sentry_${TERRAFORM_SENTRY_VERSION}_linux_amd64.zip && \
    unzip -n terraform-provider-sentry_${TERRAFORM_SENTRY_VERSION}_linux_amd64.zip -d ${TERRAFORM_PLUGINS_PATH} && \
    rm -rf /tmp/* && \
    rm -rf /var/cache/apk/* && \
    rm -rf /var/tmp/*

RUN mkdir /code
RUN mkdir /deployments
WORKDIR /code

ADD requirements.txt .
RUN pip install -r requirements.txt
COPY src /code/src/
ADD MANIFEST.in .
ADD setup.cfg . 
ADD setup.py . 
RUN python setup.py bdist_wheel && pip install dist/mach-0.0.0-py3-none-any.whl

RUN apk del .build-deps

ENTRYPOINT ["mach"]