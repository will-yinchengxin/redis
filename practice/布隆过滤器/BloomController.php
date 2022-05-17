<?php
/**
 * 布隆过滤器的简单实现
 *
 * @package   HotMai\Live\Controllers
 * @author    Will  <826895143@qq.com>
 * @copyright Copyright (C) 2021 Will
 */
namespace HotMai\Live\Controllers;

use HotMai\Live\Controllers\Base\LiveControllerBase;
use HotMai\Live\Controllers\Hash;
use HotMai\Live\Controllers\Hash;

/**
 * Class BloomController
 * @package HotMai\Live\Controllers
 */
class BloomController extends LiveControllerBase {

    /**
     * @var $bucket
     */
    protected $bucket = 'test';

    /**
     * 设置Bloom
     */
    public function setBloomAction($string = 'will') {
        $Hash = (new Hash());
        $hashFunc = $this->getHashFunc();
        $this->redis->multi();
        foreach ($hashFunc as $value) {
            $hash = $Hash::$value($string);
            $this->redis->setBit($this->bucket, $hash, 1);
        }
        $this->redis->exec();
        $this->terminalResponse(self::STATUS_SUCCESS, 'set success');
    }

    /**
     * 使用Bloom
     */
    public function judgeAction($string = 'yin') {
        $Hash = (new Hash());
        $hashFunc = $this->getHashFunc();
        $this->redis->multi();
        foreach ($hashFunc as $value) {
            $hash = $Hash::$value($string);
            $this->redis->getBit($this->bucket, $hash);
        }
        $res = $this->redis->exec();
        foreach ($res as $bit) {
            if ($bit == 0) {
                $this->terminalResponse(self::STATUS_SUCCESS, 'not exit');
            }
        }
        $this->terminalResponse(self::STATUS_SUCCESS, 'exit');
    }


    /**
     * 初始化所有方法放置于$hashFunc中
     */
    private function getHashFunc() {
        $ref = new \ReflectionClass(Hash::class);
        $methods = $ref->getMethods();
        foreach ($methods as $method) {
            if ($method->getName() != '__construct') {
                $hashFunc[] = $method->getName();
            }
        }
        return $hashFunc;
    }

}
