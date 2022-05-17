# 利用定时任务执行rdb备份
### 1.通过shell脚本定时执行rdb备份
- info persistence(查看持久化信息)
>rdb_bgsave_in_progress:0
>
>aof_rewrite_scheduled:0
>
>这两个参数为1,标识持久化在进行
````
127.0.0.1:6379> info persistence
# Persistence
loading:0
# 离最近一次成功生成rdb文件，写入命令的个数，即有多少个写入命令没有持久化
rdb_changes_since_last_save:0
# 服务器是否正在创建rdb文件
rdb_bgsave_in_progress:0
# 离最近一次成功创建rdb文件的时间戳。当前时间戳 - rdb_last_save_time=多少秒未成功生成rdb文件
rdb_last_save_time:1588400307
# 最近一次rdb持久化是否成功
rdb_last_bgsave_status:ok
rdb_last_bgsave_time_sec:0
rdb_current_bgsave_time_sec:-1
rdb_last_cow_size:167936
# 是否开启了aof
aof_enabled:1
# 标识aof的rewrite操作是否在进行中
aof_rewrite_in_progress:0
# ewrite任务计划，当客户端发送bgrewriteaof指令，如果当前rewrite子进程正在执行，那么将客户端请求的bgrewriteaof变为计划任务，待aof子进程结束后
执行rewrite
aof_rewrite_scheduled:0
aof_last_rewrite_time_sec:0
aof_current_rewrite_time_sec:-1
# 上次bgrewriteaof操作的状态
aof_last_bgrewrite_status:ok
# 上次aof写入状态
````

----
### 2.shell脚本的了解
Shell 是一个用 C 语言编写的程序，它是用户使用 Linux 的桥梁。Shell 既是一种命令语言，又是一种程序设计语言。shell 是指一种应用程序，这个应用程序提供了
一个界面，用户通过这个界面访问操作系统内核的服务。
shell脚本：简单点理解就是，和dockerfile一样把在控制台输入的命令放到一个文件中集中执行

简单体验``vim hi.sh``
````
echo "hello world shell"
````
执行``chmod +x ./hi.sh`` 使脚本具有执行权限

./hi.sh #执行脚本

----
### 3.设计思路
- 在容器中会定时执行rdb备份命令持久化redis
- 在容器中会定时执行sh脚本检测当前redis的持久化状态
- 在确定redis持久化完成之后就会把文件推送到备份的服务器
- 如果有必要在备份的服务器上会根据天适当删除一些备份的数据

---
### 4.操作须知

[crontab的使用](https://www.cnblogs.com/ftl1012/p/crontab.html)

#### 4.1 通过命令``cat /etc/crontab``查看定时任务信息
````
[root@localhost ~]# cat /etc/crontab 
SHELL=/bin/bash
PATH=/sbin:/bin:/usr/sbin:/usr/bin
MAILTO=root

# For details see man 4 crontabs

# Example of job definition:
# .---------------- minute (0 - 59)
# |  .------------- hour (0 - 23)
# |  |  .---------- day of month (1 - 31)
# |  |  |  .------- month (1 - 12) OR jan,feb,mar,apr ...
# |  |  |  |  .---- day of week (0 - 6) (Sunday=0 or 7) OR sun,mon,tue,wed,thu,fri,sat
# |  |  |  |  |
# *  *  *  *  * user-name  command to be executed
````
- 前四行是用来配置crond任务运行的环境变量
- 第一行SHELL变量指定了系统要使用哪个shell，这里是bash
- 第二行PATH变量指定了系统执行命令的路径
- 第三行MAILTO变量指定了crond的任务执行信息将通过电子邮件发送给root用户
- 如果MAILTO变量的值为空，则表示不发送任务执行信息给用户
- 第四行的HOME变量指定了在执行命令或者脚本时使用的主目录。
- 星号（*）：代表所有可能的值，如month字段为星号，则表示在满足其它字段的制约条件后每月都执行该命令操作。
- 逗号（,）：可以用逗号隔开的值指定一个列表范围，例如，“1,2,5,7,8,9”
- 中杠（-）：可以用整数之间的中杠表示一个整数范围，例如“2-6”表示“2,3,4,5,6”
- 正斜线（/）：可以用正斜线指定时间的间隔频率，例如“0-23/2”表示每两小时执行一次。
````
[root@localhost redis]# crontab -e
#min    hour    day     month   weekday command 
*/1     *       *       *       *       sh /redis/hi.sh
````
You have new mail in /var/spool/mail/root

#### 4.2 因为考虑到名称问题，这里会考虑使用时间命名
``date `+"%Y-%m-%d %H:%M.%S"``

#### 4.3 这个命令除了进入redis操作界面还可以不用进入执行命令
/redis/data # redis-cli info persistence
````
/ # redis-cli info persistence
# Persistence
loading:0
rdb_changes_since_last_save:0
rdb_bgsave_in_progress:0
````
#### 4.4 获取rdb_bgsave_in_progress内容,awk中 -F 进行查找分割
````
redis-cli info persistence | grep rdb_bgsave_in_progress | awk -F ':' '{prin
t $2}'
````
#### 4.5 ssh & scp
这是一个远程来接其他服务器的命令，并且也可以推送文件到指定的服务器
>这里使用的时apline系统,使用apk方式添加

apk add --no-cache openssh tzdata
````
验证:
/ # ssh
usage: ssh [-46AaCfGgKkMNnqsTtVvXxYy] [-B bind_interface]
           [-b bind_address] [-c cipher_spec] [-D [bind_address:]port]
           [-E log_file] [-e escape_char] [-F configfile] [-I pkcs11]
           [-i identity_file] [-J [user@]host[:port]] [-L address]
           [-l login_name] [-m mac_spec] [-O ctl_cmd] [-o option] [-p port]
           [-Q query_option] [-R address] [-S ctl_path] [-W host:port]
           [-w local_tun[:remote_tun]] destination [command]
````
#### 4.6 ssh root@192.168.100.147(连接方式)
````
The authenticity of host '192.168.100.147 (192.168.100.147)' can't be established.
ECDSA key fingerprint is SHA256:FK+UZcNjGk9Ky6yg0MZv6mPizqOOODd+svSTvQA1GO8.
Are you sure you want to continue connecting (yes/no/[fingerprint])?yes
root@192.168.100.147's password:
[root@localhost ~]# exit
````
#### 4.7 配置免密码登入操作过程如下
- ssh-keygen -t rsa(一路回车)
````
/ # ssh-keygen -t rsa
Generating public/private rsa key pair.
Enter file in which to save the key (/root/.ssh/id_rsa): 
Enter passphrase (empty for no passphrase): 
Enter same passphrase again: 
Your identification has been saved in /root/.ssh/id_rsa.
Your public key has been saved in /root/.ssh/id_rsa.pub.
The key fingerprint is:
SHA256:pKGBKmmWvf3UvI2iKDfcpCbj2FA3uFmoDcwBia+bRag root@476f1d6745cc
The key's randomart image is:
+---[RSA 3072]----+
|o.               |
|+  .             |
| +. . . .        |
|+o*o o +         |
|*O+.= . S        |
|E=.=oo  o        |
|o+=.+. . o       |
|+* *..o.  +      |
|o.B......o .     |
+----[SHA256]-----+
````
````
/ # cd /root/.ssh/
~/.ssh # ls
id_rsa //私钥      id_rsa.pub//公钥   known_hosts
````
将本机的公钥复制到远程机器的authorized_keys文件中(通过ssh-copy-id/scp)
>这里我们采用ssh-copy-id

- ssh-copy-id -i ~/.ssh/id_rsa.pub 192.168.100.147
- scp -p ~/.ssh/id_rsa.pub root@192.168.100.147:/root/.ssh
````
~/.ssh # ssh-copy-id -i ~/.ssh/id_rsa.pub 192.168.100.147
/usr/bin/ssh-copy-id: INFO: Source of key(s) to be installed: "/root/.ssh/id_rsa.pub"
/usr/bin/ssh-copy-id: INFO: attempting to log in with the new key(s), to filter out any that are already installed
expr: warning: '^ERROR: ': using '^' as the first character
of a basic regular expression is not portable; it is ignored
/usr/bin/ssh-copy-id: INFO: 1 key(s) remain to be installed -- if you are prompted now it is to install the new keys
root@192.168.100.147's password: 

Number of key(s) added: 1

Now try logging into the machine, with:   "ssh '192.168.100.147'"
and check to make sure that only the key(s) you wanted were added.
````
此时测试,就可以免密登录
````
~/.ssh # ssh root@192.168.100.147
Last login: Sat May 16 08:05:02 2020 from 192.168.100.146
[root@localhost ~]# 
````
#### 4.8 通过 scp 将内容写到对方的文件中
scp -p ~/.ssh/id_rsa.pub root@192.168.169.150:/redis/data/
````
~/.ssh # scp -p ~/.ssh/id_rsa.pub 192.168.100.147:/redis/data
id_rsa.pub                                    100%  571   583.2KB/s   00:00

~/.ssh # ssh 192.168.100.147
Last login: Sat May 16 08:24:36 2020 from 192.168.100.146

[root@localhost ~]# cd /redis/data
[root@localhost data]# ll
total 4
-rw-r--r--. 1 root root 571 May  3 13:43 id_rsa.pub
````

---
### 5.[脚本的书写](https://blog.csdn.net/sheqianweilong/article/details/88833275)
- rdb实时备份
````
之所以能用到xargs，关键是由于很多命令不支持|管道来传递参数，而日常工作中有有这个必要
例如
find /redis/data/*rdb | rm -rf       #这个命令是错误的
find /redis/data/*rdb | xargs rm -rf   #这样才是正确的
````
````
msg=`redis-cli bgsave`
result=`redis-cli info Persistence |grep rdb_bgsave_in_progress |awk -F":" '{print $2}'`
while [ `echo $result` -eq "1" ] ;
do
sleep 1
result=`redis-cli info Persistence |grep rdb_bgsave_in_progress |awk -F":" '{print $2}'`
done
dateDir=`date +%Y%m%d%H`
dateFile=`date +%M`
scpData=/redis/data/rdb/$dateDir
mv /redis/data/dump.rdb /redis/data/$dateFile".rdb"
ssh root@192.168.169.150 "mkdir -p "$scpData
scp -p /redis/data/$dateFile".rdb" root@192.168.169.150:$scpData
find /redis/data/*rdb | xargs rm -rf
````
- aof实时备份
实现一下aof的备份；要求根据时间删除备份服务器的前两天的数据，以分钟作为备份
````
[root@localhost aof]# vim aof.sh  
result=`redis-cli info Persistence |grep aof_rewrite_scheduled |awk -F":" '{print $2}'`
while [ `echo $result` -eq  1  ];
do
        sleep 1
        result=`redis-cli info Persistence |grep aof_rewrite_scheduled |awk -F":" '{print $2}'`
done
dateDir=`date +%Y%m%d`
dateFile=`date +%M`
scpData=/redis-sh/aof/$dateDir
ssh root@192.168.100.147 "mkdir -p "$scpData
scp -p /www/server/redis/appendonly.aof  192.168.100.147:$scpData/$dateFile".aof"
ssh root@192.168.100.147 "find /redis-sh/aof -ctime/mtime/mmin/cmin +2 -name "*" -exec rm -rf {} \;"
````
定时器内容(crontab -e):
````
*/1 * * * *  sh /redis-sh/aof/aof.sh
````