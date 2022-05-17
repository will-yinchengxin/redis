CREATE TABLE `red_packet_info` (
    `id` int(11) NOT NULL AUTO_INCREMENT,
    `red_packet_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '红包id，采⽤\r\n	timestamp+5位随机数',
    `total_amount` int(11) NOT NULL DEFAULT '0' COMMENT '红包总⾦额，单位分',
    `total_packet` int(11) NOT NULL DEFAULT '0' COMMENT '红包总个数',
    `remaining_amount` int(11) NOT NULL DEFAULT '0' COMMENT '剩余红包⾦额，单位\r\n	分',
    `remaining_packet` int(11) NOT NULL DEFAULT '0' COMMENT '剩余红包个数',
    `uid` int(11) NOT NULL DEFAULT '0' COMMENT '新建红包⽤户的⽤户标识',
    `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COMMENT='红包信息\r\n表，新建⼀个红包插⼊⼀条记录'


CREATE TABLE `red_packet_record` (
    `id` int(11) NOT NULL AUTO_INCREMENT,
    `amount` int(11) NOT NULL DEFAULT '0' COMMENT '抢到红包的⾦额',
    `nick_name` varchar(32) NOT NULL DEFAULT '0' COMMENT '抢到红包的⽤户的⽤户\r\n名',
    `img_url` varchar(255) NOT NULL DEFAULT '0' COMMENT '抢到红包的⽤户的头像',
    `uid` int(11) NOT NULL DEFAULT '0' COMMENT '抢到红包⽤户的⽤户标识',
    `red_packet_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '红包id，采⽤\r\ntimestamp+5位随机数',
    `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COMMENT='抢红包记\r\n录表，抢⼀个红包插⼊⼀条记录'