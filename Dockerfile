#   docker build -t mefs-test .

#   docker run -it -v D:\docker\test0\nodetest:/root/.mefs -e MEFSROLE=provider --entrypoint=/bin/bash mefs-test

FROM limcos/environment_construction:latest

MAINTAINER suzakinishi <ccyansnow@gmail.com>

# download mefs

RUN wget -P /usr/local/bin/ http://212.64.28.207:4000/mefs    \
 && chmod 777 /usr/local/bin/mefs   

# 4001: Swarm TCP; should be exposed to the public
# 5001: Daemon API; must not be exposed publicly but to client services under you control

EXPOSE 4001
EXPOSE 5001