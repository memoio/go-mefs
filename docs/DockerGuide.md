# docker环境搭建与镜像，容器运行

## ubuntu安装docker

其他环境参考<https://yeasy.gitbooks.io/docker_practice/install/>  

```bash
#install curl
apt-get update
apt-get install -y curl

#install-docker
curl -fsSL get.docker.com -o get-docker.sh
sh get-docker.sh --mirror Aliyun  

#install speeder
curl -sSL https://get.daocloud.io/daotools/set_mirror.sh | sh -s http://573b0ee5.m.daocloud.io  
systemctl restart docker.service  

#test docker
docker run hello-world
```

## 镜像搭建

可能需要上传最新的mefs二进制

```bash
scp -P 26186 $GOPATH/bin/mefs mefs@97.64.124.20:~/bin
```

```docker
docker build -t [镜像名] .
```

## 生成容器

```docker
docker run -itd -v [你希望的ipfs文件储存路径]:/root/.mefs -e MEFSROLE=[user/keeper/provider] [镜像名]
```

常用可用选项如下

```
-e 环境变量修改 不输的话则默认值为user
-it 前台运行
-itd 后台运行
--rm 容器停止后自动删除
--entrypoint=/bin/bash 使容器运行时，忽略镜像中的CMD命令，直接启用bash
-v 本地目录映射 你希望的ipfs文件储存路径（windows系统路径参考写法 D:\nodetest）:容器里面的路径
```

## Docker常用命令

```docker
docker image ls     列出镜像
docker image rm [选项] <镜像1> [<镜像2> ...]     删除镜像
docer ps -a      列出所有容器，去掉-a则是列出运行容器
docker container ls     列出容器
docker container stop [容器名]     终止运行中的容器
docker container start [容器名]     重启终止的容器
docker container restart [容器名]     重启运行中的容器
docker container rm [容器名]     删除一个处于终止状态的容器
docker container prune     删除所有处于终止状态的容器
docker exec -it [容器名] [操作名]     进入后台运行的容器进行操作
docker logs [选项] [容器名]     打印后台运行的容器的log
容器退出时如果想继续在后台运行：按顺序按[ctrl+p][ctrl+q]，如果不想继续运行：按[ctrl+d]或输入exit
```