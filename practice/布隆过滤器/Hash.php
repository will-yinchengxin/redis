<?php
/**
 * Hash算法集合
 *
 * @package   HotMai\Live\Controllers
 * @author    Will  <826895143@qq.com>
 * @copyright Copyright (C) 2021 Will
 */
namespace HotMai\Live\Controllers;

/**
 * Class Hash
 * @package HotMai\Live\Controllers
 */
class Hash {

    /**
     * 由Justin Sobel编写的按位散列函数
     */
    public static function JSHash($string, $len = null) {
        $hash = 1315423911;
        $len || $len = strlen($string);
        for ($i = 0; $i < $len; $i++) {
            // ^ 两个位相同为0，相异为1,
            $hash ^= (($hash << 5) + ord($string[$i]) + ($hash >> 2));
        }
        // & 两个位都为1时，结果才为1
        // 0xFFFFFFFF代表-1
        return ($hash % 0xFFFFFFFF) & 0xFFFFFFFF;
    }

    /**
     * 该hash算法基于AT & T 贝尔实验室的Peter J. Weinberger的算法成果
     */
    public static function PJWHash($string, $len = null) {
        $bitsInUnsignedInt = 4 * 8; // unsigned int
        $threeQuarters = ($bitsInUnsignedInt * 3) / 4;
        $oneEighth = $bitsInUnsignedInt / 8;
        $highBits = 0xFFFFFFFF << intval($bitsInUnsignedInt - $oneEighth);
        $hash = 0;
        $len || $len = strlen($string);
        for ($i = 0; $i < $len; $i++) {
            $hash = ($hash << intval($oneEighth)) + ord($string[$i]);
        }
        $test = $hash & $highBits;
        if ($test != 0) {
            // ~ 0变1，1变0
            $hash = (($hash ^($test >> intval($threeQuarters))) & (~$highBits));
        }
        return ($hash % 0xFFFFFFFF) & 0xFFFFFFFF;
    }

    /**
     * 类似于PJWHash功能，但针对32位处理器进行了调整。它是基于UNIX的系统上的widley使用哈希函数。
     */
    public static function ELFHash($string, $len = null) {
        $hash = 0;
        $len || $len = strlen($string);
        for ($i = 0; $i < $len; $i++) {
            $hash = ($hash << 4) + ord($string[$i]);
            $x = $hash & 0xF0000000;
            if ($x != 0) {
                $hash ^= ($x >> 24);
            }
            $hash &= ~$x;
        }
        return ($hash % 0xFFFFFFFF) & 0xFFFFFFFF;
    }

    /**
     * 这是在开源SDBM项目中使用的首选算法。
     * 哈希函数似乎对许多不同的数据集具有良好的总体分布。它似乎适用于数据集中元素的MSB存在高差异的情况。
     */
    public static function SDBHash($string, $len = null) {
        $hash = 0;
        $len || $len = strlen($string);
        for ($i = 0; $i < $len; $i++) {
            $hash = (int) (ord($string[$i]) + ($hash << 6) + ($hash << 16) - $hash);
        }
        return ($hash % 0xFFFFFFFF) & 0xFFFFFFFF;
    }

}
