FROM alpine:3.11
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
  && apk add  gcc g++ libc-dev  wget vim  openssl-dev make  linux-headers \
  && rm -rf /var/cache/apk/*

COPY ./redis-5.0.7.tar.gz redis-5.0.7.tar.gz

RUN mkdir -p /usr/src/redis \
      && mkdir -p /redis/data \
      && mkdir -p /redis/conf \
      && mkdir -p /redis/log   \
      && mkdir -p /var/log/redis \
      && tar -zxvf redis-5.0.7.tar.gz -C /usr/src/redis \
      && rm -rf redis-5.0.7.tar.gz \
      && cd /usr/src/redis/redis-5.0.7 && make \
      && cd /usr/src/redis/redis-5.0.7 && make install;

EXPOSE 6379

CMD ["redis-server","/usr/src/redis/redis-5.0.7/redis.conf"]