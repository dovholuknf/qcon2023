###################################################
#
# CHANGE THESE AS YOU SEE FIT
#
SOURCE_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
TMP_DIR=/tmp/dovholuknf/qcon2023
SPIRE_VERSION=1.6.4
SPIRE_CMD=${TMP_DIR}/spire-${SPIRE_VERSION}/bin/spire-server
OPENZITI_VER=0.28.0
#
###################################################
echo "------------------------------------------"
echo SOURCE_DIR=${SOURCE_DIR}
echo TMP_DIR=${TMP_DIR}
echo SPIRE_VERSION=${SPIRE_VERSION}
echo SPIRE_CMD=${SPIRE_CMD}
echo "------------------------------------------"

echo "recreating TMP_DIR: $TMP_DIR"
sudo rm -rf ${TMP_DIR}
mkdir -p ${TMP_DIR}

cd ${SOURCE_DIR}/src
echo "compiling all the samples with go build..."
echo "nosecurity-*"
go build -o ${TMP_DIR}/nosecurity-server part1_nosecurity/server/main.go
go build -o ${TMP_DIR}/nosecurity-client part1_nosecurity/client/main.go

echo "spire-*"
go build -o ${TMP_DIR}/spire-server part2_spire/server/main.go
go build -o ${TMP_DIR}/spire-client part2_spire/client/main.go

echo "openziti-*"
go build -o ${TMP_DIR}/openziti-server part3_openziti/server/main.go
go build -o ${TMP_DIR}/openziti-client part3_openziti/client/main.go

echo "spire-and-openziti-*"
go build -o ${TMP_DIR}/spire-and-openziti-server part4_spire_and_openziti/server/main.go
go build -o ${TMP_DIR}/spire-and-openziti-client part4_spire_and_openziti/client/main.go

sudo killall spire-server
sudo killall spire-agent
sudo killall oidc-discovery-provider

cd ${TMP_DIR}
echo "downloading and untarring SPIRE from https://github.com/spiffe/spire/releases/download/v${SPIRE_VERSION}/spire-${SPIRE_VERSION}-linux-amd64-glibc.tar.gz"
curl -s -N -L https://github.com/spiffe/spire/releases/download/v${SPIRE_VERSION}/spire-${SPIRE_VERSION}-linux-amd64-glibc.tar.gz | tar xz
curl -s -N -L https://github.com/spiffe/spire/releases/download/v${SPIRE_VERSION}/spire-extras-${SPIRE_VERSION}-linux-amd64-glibc.tar.gz | tar xz
mv spire-extras-${SPIRE_VERSION}/bin/oidc-discovery-provider spire-${SPIRE_VERSION}/bin/
mv spire-extras-${SPIRE_VERSION}/conf/oidc-discovery-provider spire-${SPIRE_VERSION}/conf

cd ${TMP_DIR}/spire-${SPIRE_VERSION}/
echo "emitting configuration file for SPIRE server"
sed -i -e 's/bind_address = .*/bind_address = "0.0.0.0"/g' conf/server/server.conf
sed -i -e 's/bind_port = .*/bind_port = "8600"/g' conf/server/server.conf
sed -i -e 's/trust_domain = .*/trust_domain = "openziti"/g' conf/server/server.conf
sed -i -e 's/"48h"/"48h"\n    "jwt_issuer" = "zpire"/g' conf/server/server.conf


#start spire...
$SPIRE_CMD run -config conf/server/server.conf > $TMP_DIR/spire.server.log &
echo "spire-server started. waiting two seconds for it to start up and initialize..."
sleep 2

#setup spire agent
agent_token=$(${SPIRE_CMD} token generate -spiffeID  spiffe://openziti/ids | cut -d " " -f2)
echo "SPIRE AGENT TOKEN: $agent_token"


echo "emitting configuration file for SPIRE agent"
cat > conf/agent/agent.conf << HERE
agent {
    data_dir = "./data/agent"
    log_level = "INFO"
    trust_domain = "openziti"
    server_address = "localhost"
    server_port = "8600"

    # Insecure bootstrap is NOT appropriate for production use but is ok for
    # simple testing/evaluation purposes.
    insecure_bootstrap = true
}

plugins {
   KeyManager "disk" {
        plugin_data {
            directory = "./data/agent"
        }
    }

    NodeAttestor "join_token" {
        plugin_data {}
    }

    WorkloadAttestor "unix" {
        plugin_data {
            discover_workload_path = true
        }
    }
}
HERE

sudo -b bin/spire-agent run -config conf/agent/agent.conf -joinToken $agent_token > $TMP_DIR/spire.agent.log
echo "spire-agent started. waiting two seconds for it to start up and initialize..."
sleep 2

mkdir -p conf/oidc-discovery-provider/

echo "emitting configuration file for SPIRE oidc-discovery-provider extra"
cat > conf/oidc-discovery-provider/oidc-discovery-provider.conf << HERE
log_level = "debug"
# allowed domain list, I keep mypublicdomain.test here just for displaying it is a list,
# but since it is for local running, you will be using only localhost
domains = ["mypublicdomain.test", "localhost"]
# Suggested only for local environments
allow_insecure_scheme = true
insecure_addr = ":8601"
server_api {
    address = "unix:///tmp/spire-server/private/api.sock"
}
HERE
#start oidc keys
bin/oidc-discovery-provider -config conf/oidc-discovery-provider/oidc-discovery-provider.conf > $TMP_DIR/spire.oidc.discover.provider.log &

echo "SPIRE setup complete. Registering expected workloads"

$SPIRE_CMD entry create \
  -spiffeID spiffe://openziti/jwtServer \
  -parentID spiffe://openziti/ids \
  -dns openziti.ziti \
  -dns openziti.spire.ziti \
  -dns localhost \
  -selector unix:path:${TMP_DIR}/spire-server

$SPIRE_CMD entry create \
  -spiffeID spiffe://openziti/jwtServer \
  -parentID spiffe://openziti/ids \
  -dns openziti.ziti \
  -dns openziti.spire.ziti \
  -dns localhost \
  -selector unix:path:${TMP_DIR}/openziti-server

$SPIRE_CMD entry create \
  -spiffeID spiffe://openziti/jwtServer \
  -parentID spiffe://openziti/ids \
  -dns openziti.ziti \
  -dns openziti.spire.ziti \
  -dns localhost \
  -selector unix:path:${TMP_DIR}/spire-and-openziti-server

$SPIRE_CMD entry create \
  -spiffeID spiffe://openziti/jwtClient \
  -parentID spiffe://openziti/ids \
  -selector unix:path:${TMP_DIR}/spire-client

$SPIRE_CMD entry create \
  -spiffeID spiffe://openziti/jwtClient \
  -parentID spiffe://openziti/ids \
  -selector unix:path:${TMP_DIR}/openziti-client

$SPIRE_CMD entry create \
  -spiffeID spiffe://openziti/jwtClient \
  -parentID spiffe://openziti/ids \
  -selector unix:path:${TMP_DIR}/spire-and-openziti-client


echo "starting OpenZiti environment via docker compose -d"
curl -s https://get.openziti.io/dock/docker-compose.yml > $TMP_DIR/docker-compose.yml
curl -s https://get.openziti.io/dock/.env > $TMP_DIR/.env

docker compose -f $TMP_DIR/docker-compose.yml --env-file=$TMP_DIR/.env -p qcon2023 down -v
docker compose -f $TMP_DIR/docker-compose.yml --env-file=$TMP_DIR/.env -p qcon2023 up -d

ziti_ctrl="https://localhost:1280"
while [[ "$(curl -w "%{http_code}" -m 1 -s -k -o /dev/null ${ziti_ctrl}/version)" != "200" ]]; do echo "waiting for ${ziti_ctrl}"; sleep 3; done; echo "controller online"


eval $(docker exec qcon2023-ziti-controller-1 cat ziti.env | grep ZITI_PWD=)

echo "getting openziti"
curl -s -N -L https://github.com/openziti/ziti/releases/download/v${OPENZITI_VER}/ziti-linux-amd64-${OPENZITI_VER}.tar.gz | tar xz -C ${TMP_DIR}

${TMP_DIR}/ziti/ziti edge login $ziti_ctrl -u admin -p $ZITI_PWD -y
echo "logged into ziti..."

${TMP_DIR}/ziti/ziti edge delete service-policy demo-services-bind-policy
${TMP_DIR}/ziti/ziti edge delete service-policy demo-services-dial-policy
${TMP_DIR}/ziti/ziti edge delete config openziti-and-spire-intercept.v1
${TMP_DIR}/ziti/ziti edge delete service openziti-and-spire-service
${TMP_DIR}/ziti/ziti edge delete config openziti-only-intercept.v1
${TMP_DIR}/ziti/ziti edge delete service openziti-only-service
${TMP_DIR}/ziti/ziti edge delete identity zpire-jwtClient
${TMP_DIR}/ziti/ziti edge delete identity zpire-jwtServer
${TMP_DIR}/ziti/ziti edge delete auth-policy zpire-auth-policy
${TMP_DIR}/ziti/ziti edge delete ext-jwt-signer zpire-ext-jwt


signer=$(${TMP_DIR}/ziti/ziti edge create ext-jwt-signer zpire-ext-jwt zpire -u http://172.20.166.120:8601/keys -a "spiffe://openziti/jwtServer")
authPolicy=$(${TMP_DIR}/ziti/ziti edge create auth-policy zpire-auth-policy --primary-ext-jwt-allowed --primary-ext-jwt-allowed-signers ${signer})

# create two identities for 'server' and 'clients'
${TMP_DIR}/ziti/ziti edge create identity service zpire-jwtClient \
  --auth-policy $authPolicy \
  --external-id "spiffe://openziti/jwtClient" \
  -a demo-services-client
${TMP_DIR}/ziti/ziti edge create identity service zpire-jwtServer \
  --auth-policy $authPolicy \
  --external-id "spiffe://openziti/jwtServer" \
  -a demo-services-server

# create two demo services
${TMP_DIR}/ziti/ziti edge create config openziti-only-intercept.v1 intercept.v1 \
  '{"protocols":["tcp"],"addresses":["openziti.ziti"], "portRanges":[{"low":443, "high":443}]}'
${TMP_DIR}/ziti/ziti edge create service openziti-only-service --configs openziti-only-intercept.v1 -a demo-services

${TMP_DIR}/ziti/ziti edge create config openziti-and-spire-intercept.v1 intercept.v1 \
  '{"protocols":["tcp"],"addresses":["openziti.spire.ziti"], "portRanges":[{"low":443, "high":443}]}'
${TMP_DIR}/ziti/ziti edge create service openziti-and-spire-service --configs openziti-and-spire-intercept.v1 -a demo-services

# authorize identities to dial/bind
${TMP_DIR}/ziti/ziti edge create service-policy demo-services-bind-policy Bind \
  --service-roles '#demo-services' \
  --identity-roles '#demo-services-server'
${TMP_DIR}/ziti/ziti edge create service-policy demo-services-dial-policy Dial \
  --service-roles '#demo-services' \
  --identity-roles '#demo-services-client'

${TMP_DIR}/ziti/ziti edge create identity user local.docker.user -a demo-services-client -o ${TMP_DIR}/local.docker.user.jwt

echo "ziti configuration applied. test identity for tunneler at: ${TMP_DIR}/local.docker.user.jwt"
echo " "
echo "At this point you should be able to run any of the servers and clients:"
echo " - ${TMP_DIR}/nosecurity-server"
echo " - ${TMP_DIR}/nosecurity-client 10 \"*\" 2"
echo " "
echo " - ${TMP_DIR}/spire-server"
echo " - ${TMP_DIR}/spire-client 10 \"*\" 2"
echo " "
echo " - ${TMP_DIR}/openziti-server"
echo " - ${TMP_DIR}/openziti-client 10 \"*\" 2"
echo " "
echo " - ${TMP_DIR}/spire-and-openziti-server"
echo " - ${TMP_DIR}/spire-and-openziti-client 10 \"*\" 2"
echo " "