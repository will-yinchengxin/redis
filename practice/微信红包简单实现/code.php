<?php
/**
 * 微信红包简单实现
 *
 * @package   HotMai\Live\Controllers
 * @author    Will  <826895143@qq.com>
 * @copyright Copyright (C) 2021 Will
 */
namespace HotMai\Live\Controllers;

use HotMai\Base\Models\RedPacketInfo;
use HotMai\Base\Models\RedPacketRecord;
use HotMai\Live\Controllers\Base\LiveControllerBase;

/**
 * Class RedController
 * @package HotMai\Live\Controllers
 */
class RedController extends LiveControllerBase {

    // 红包个数及总金额我们給死
    protected $total = 10;
    // 金额方便计数,这里*100,以为最小金额为0.01
    protected $money = 100*100;

    /**
     * send the red pocket
     */
    public function sendAction() {
        // 将红包信息存储至数据库
        $redPackInfo = (new RedPacketInfo());
        $packId = strval(time()).strval(rand(1000,9999));
        $redPackInfo->setRedPacketId($packId); // 这里最好雪花算法进行设置
        // 红包总 个数/金额 及剩余 个数/金额
        $redPackInfo->setTotalPacket($this->total);
        $redPackInfo->setRemainingPacket($this->total);
        $redPackInfo->setTotalAmount($this->money);
        $redPackInfo->setRemainingAmount($this->money);
        $redPackInfo->setUid($this->getLoginSubscriberId()); // 获取发动红包者的用户id

        if ($redPackInfo->save()) {
            // 将红包总数及金额总数存储至redis中
            $this->redis->set(sprintf(self::TOTAL_PACKET, $packId, $this->total), $this->total);
            $this->redis->set(sprintf(self::TOTAL_AMOUNT, $packId, $this->money), $this->money);
            $this->terminalResponse(self::STATUS_SUCCESS, 'success', ['packId' => $packId]);
        }
    }

    /**
     * rab the red pocket
     *
     * 抢红包功能属于原子减操作
     * 当大小小于 0 时原子减失败
     * 当红包个数为0时，后面进来的用户全部抢红包失败，并不会进入拆红包环节
     *
     * 将红包ID的请求放入请求队列中，如果发现超过红包的个数，直接返回
     * 抢到红包不一定能拆成功
     *
     * 每个分得金额的计算方式: (剩余总金额/剩余总人数)*2, 那么每个人所获金额为 0.01~(剩余总金额/剩余总人数)*2 之间
     */
    public function robAction() {
        // 假设上一步获取的红包id为 16267656461631
        $packId = 16267720757050;
        // 红包总数及金额我们这里手动给出,根据业务自行获取
        $total = 10;
        $amount = $this->money;
        // redis中存储红包个数的键
        $packTotalCount = sprintf(self::TOTAL_PACKET, $packId, $total);
        // redis中存储红包金额的键
        $packTotalAmount = sprintf(self::TOTAL_AMOUNT, $packId, $amount);
        if ($this->redis->exists($packTotalCount)) {
            // 红包剩余总个数
            $packNumValue = $this->redis->get($packTotalCount);
            if ($packNumValue != null && $packNumValue > 0) {
                // 红包个数减一
                $this->redis->decr($packTotalCount);

                if ($this->redis->exists($packTotalAmount)) {
                    // 红包剩余金额总值
                    $packAmountValue = $this->redis->get($packTotalAmount);

                    // 可获得的最大金额数
                    $maxMoney = ($packAmountValue/$packNumValue)*2;
                    // 获取随机的金额
                    var_dump($packAmountValue, $max = $this->getRandMoney($maxMoney));
                    $randMoney = ($packNumValue == 1 ? $packAmountValue : $max);
                    $this->redis->decrBy($packTotalAmount, $randMoney);

                    // 更新数据库中的数据
                    // 这里用户id使用随机数代替
                    $uid = mt_rand(1000, 9999);
                    if ($this->updateDBInfo($packId, $randMoney, $uid)) {
                        $this->terminalResponse(self::STATUS_SUCCESS, sprintf('抢到金额%s', $randMoney ?? 0));
                    }
                    $this->terminalResponse(self::STATUS_FAILURE, '操作有误!');
                }
            } else {
                $this->terminalResponse(self::STATUS_FAILURE, '红包被抢光了');
            }
        } else {
            $this->terminalResponse(self::STATUS_FAILURE, '红包不存在!');
        }
    }

    /**
     * 更新数据库内容
     */
    public function updateDBInfo($packId, $randMoney, $uid) {
        try {
            $this->db->begin();
            // 总库中处理金额的减少
            $RedPacketInfo = RedPacketInfo::findFirst(array(
                'conditions' => 'red_packet_id = ?0 ',
                'bind' => [$packId]
            )) ?: null;
            $Amount = $RedPacketInfo->getRemainingAmount();
            $Packet = $RedPacketInfo->getRemainingPacket();
            if ($Amount > $randMoney) {
                $RedPacketInfo->setRemainingAmount($Amount - $randMoney);
            } else {
                $RedPacketInfo->setRemainingAmount(0);
            }
            $RedPacketInfo->setRemainingPacket($Packet - 1);
            // 分库中记录每一条详细信息
            $RedPacketRecord = (new RedPacketRecord);
            $RedPacketRecord->setUid($uid);
            $RedPacketRecord->setRedPacketId($packId);
            $RedPacketRecord->setAmount($randMoney);
            // Todo 更为详细的信息存储
            $RedPacketInfo->save();
            $RedPacketRecord->save();
            $this->db->commit();
            return true;
        } catch (\Exception $e) {
            $this->logger->error("添加关键词监控任务 获取信息错误原因:{$e->getMessage()}");
            $this->db->rollback();
            return false;
        }
    }

    /**
     * 获取随机金额
     *
     * @param int|float $min
     * @param int $max
     * @return float
     */
    public function getRandMoney(int $max, $min = 1): float {
        // 返回一个随机小数
        // return round($min +  abs($max - $min) * mt_rand(0,mt_getrandmax())/mt_getrandmax(), 2);
        return rand($max, $min);
    }

    private const TOTAL_PACKET = 'Packet_%s_TotalPacket_%s';
    private const TOTAL_AMOUNT = 'Packet_%s_TotalAmount_%s';
}
