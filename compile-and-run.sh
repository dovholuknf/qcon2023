###################################################
#
# CHANGE THESE AS YOU SEE FIT
#
SOURCE_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
TMP_DIR=/tmp/dovholuknf/qcon2023
SPIRE_VERSION=1.6.4
#
###################################################

SPIRE=${TMP_DIR}/spire-${SPIRE_VERSION}/bin/spire-server
sudo rm -rf ${TMP_DIR}
mkdir -p ${TMP_DIR}

sudo killall spire-server
sudo killall spire-agent
sudo killall oidc-discovery-provider

cd ${TMP_DIR}
curl -s -N -L https://github.com/spiffe/spire/releases/download/v${SPIRE_VERSION}/spire-${SPIRE_VERSION}-linux-amd64-glibc.tar.gz | tar xz
curl -s -N -L https://github.com/spiffe/spire/releases/download/v${SPIRE_VERSION}/spire-extras-${SPIRE_VERSION}-linux-amd64-glibc.tar.gz | tar xz
mv spire-extras-${SPIRE_VERSION}/bin/oidc-discovery-provider spire-${SPIRE_VERSION}/bin/
mv spire-extras-${SPIRE_VERSION}/conf/oidc-discovery-provider spire-${SPIRE_VERSION}/conf

cd ${TMP_DIR}/spire-${SPIRE_VERSION}/
sed -i -e 's/bind_address = .*/bind_address = "0.0.0.0"/g' conf/server/server.conf
sed -i -e 's/bind_port = .*/bind_port = "8600"/g' conf/server/server.conf
sed -i -e 's/trust_domain = .*/trust_domain = "openziti"/g' conf/server/server.conf
sed -i -e 's/"48h"/"48h"\n    "jwt_issuer" = "zpire"/g' conf/server/server.conf

#start spire...
$SPIRE run -config conf/server/server.conf > spire.server.log &
echo "spire-server started. waiting two seconds for it to start up and initialize..."
sleep 2

#setup spire agent
agent_token=$($SPIRE token generate -spiffeID  spiffe://openziti/ids | cut -d " " -f2)
#client_agent_token=$($SPIRE token generate -spiffeID  spiffe://openziti/ids | cut -d " " -f2)

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

#start the spire agent
sudo bin/spire-agent run -config conf/agent/agent.conf -joinToken $agent_token > spire.agent.log &
echo "spire-agent started. waiting two seconds for it to start up and initialize..."
sleep 2

mkdir -p conf/oidc-discovery-provider/

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
bin/oidc-discovery-provider -config conf/oidc-discovery-provider/oidc-discovery-provider.conf > spire.oidc.discover.provider.log &

# all the samples...


#$SPIRE entry create \
#  -spiffeID spiffe://openziti/jwtServer \
#  -parentID spiffe://openziti/ids \
#  -selector unix:user:cd

$SPIRE entry create \
  -spiffeID spiffe://openziti/jwtServer \
  -parentID spiffe://openziti/ids \
  -selector unix:path:/mnt/v/work/git/github/dovholuknf/qcon2023/src/linux-build/server

$SPIRE entry create \
  -spiffeID spiffe://openziti/jwtServer \
  -parentID spiffe://openziti/ids \
  -selector unix:path:/home/cd/golang-demo/server

$SPIRE entry create \
  -spiffeID spiffe://openziti/jwtServer \
  -parentID spiffe://openziti/ids \
  -selector unix:user:spire-server-workload


#$SPIRE entry create \
#  -spiffeID spiffe://openziti/jwtClient \
#  -parentID spiffe://openziti/ids \
#  -selector unix:user:spire-client-workload
$SPIRE entry create \
  -spiffeID spiffe://openziti/jwtClient \
  -parentID spiffe://openziti/ids \
  -selector unix:path:/mnt/v/work/git/github/dovholuknf/qcon/go-spiffe/v2/examples/spiffe-jwt/linux-build/client
$SPIRE entry create \
  -spiffeID spiffe://openziti/jwtClient \
  -parentID spiffe://openziti/ids \
  -selector unix:path:/home/cd/golang-demo/client
$SPIRE entry create \
  -spiffeID spiffe://openziti/jwtClient \
  -parentID spiffe://openziti/ids \
  -selector unix:user:cd

$SPIRE entry create \
  -spiffeID spiffe://openziti/jwtClient \
  -parentID spiffe://openziti/ids \
  -selector unix:path:/mnt/v/work/git/github/dovholuknf/qcon2023/src/linux-build/client

function deleteSvidBySelector {
$SPIRE entry show -selector $1 | grep Entry | cut -d ":" -f2 | xargs $SPIRE entry delete -entryID
}


ziti_ctrl="https://localhost:1280"
while [[ "$(curl -w "%{http_code}" -m 1 -s -k -o /dev/null ${ziti_ctrl}/version)" != "200" ]]; do echo "waiting for ${ziti_ctrl}"; sleep 3; done; echo "controller online"

eval $(docker exec compose-ziti-controller-1 cat ziti.env | grep ZITI_PWD=)

ziti edge login $ziti_ctrl -u admin -p $ZITI_PWD -y
echo "logged into ziti..."

ziti edge delete service-policy secure-service-binder
ziti edge delete service-policy secure-service-dialer
ziti edge delete service secure-service
ziti edge delete config secure-service-intercept.v1
ziti edge delete identity zpire-jwtClient
ziti edge delete identity zpire-jwtServer
ziti edge delete auth-policy zpire-auth-policy
ziti edge delete ext-jwt-signer zpire-ext-jwt


signer=$(ziti edge create ext-jwt-signer zpire-ext-jwt zpire -u http://172.20.166.120:8601/keys -a "spiffe://openziti/jwtServer")
authPolicy=$(ziti edge create auth-policy zpire-auth-policy --primary-ext-jwt-allowed --primary-ext-jwt-allowed-signers ${signer})
ziti edge create identity service zpire-jwtClient --auth-policy $authPolicy --external-id "spiffe://openziti/jwtClient" -a secure-service-dialers
ziti edge create identity service zpire-jwtServer --auth-policy $authPolicy --external-id "spiffe://openziti/jwtServer" -a secure-service-binders
ziti edge create config secure-service-intercept.v1 intercept.v1 '{"protocols":["tcp"],"addresses":["jwt.local.server"], "portRanges":[{"low":443, "high":443}]}'
ziti edge create service secure-service --configs secure-service-intercept.v1 -a secure-service-binders
ziti edge create service-policy secure-service-binder Bind --service-roles '@secure-service' --identity-roles '#secure-service-binders'
ziti edge create service-policy secure-service-dialer Dial --service-roles '@secure-service' --identity-roles '#secure-service-dialers'

ziti edge create identity user local.docker.user -a secure-service-dialers -o /mnt/v/temp/local.docker.user.jwt

echo "ziti configuration applied"
echo " "



