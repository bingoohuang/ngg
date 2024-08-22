package ss

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
)

// RandProvinceAndCity 返回随机省/城市
func RandProvinceAndCity() string {
	return ProvinceCity[RandIntBetween(0, len(ProvinceCity))]
}

// RandAddress 返回随机地址
func RandAddress() string {
	return RandProvinceAndCity() +
		RandChinese(2, 3) + "路" +
		strconv.Itoa(RandIntBetween(1, 8000)) + "号" +
		RandChinese(2, 3) + "小区" +
		strconv.Itoa(RandIntBetween(1, 20)) + "单元" +
		strconv.Itoa(RandIntBetween(101, 2500)) + "室"
}

// RandBankNo 返回随机银行卡号，银行卡号符合LUHN 算法并且有正确的卡 bin 前缀
func RandBankNo() string {
	// 随机选中银行卡卡头
	bank := CardBins[RandIntn(len(CardBins))]
	// 获取 卡前缀(cardBin)
	prefixes := bank.Prefixes
	// 获取当前银行卡正确长度
	cardNoLength := bank.Length
	// 生成 长度-1 位卡号
	preCardNo := strconv.Itoa(prefixes[RandIntn(len(prefixes))]) +
		fmt.Sprintf("%0*d", cardNoLength-7, RandInt64n(int64(math.Pow10(cardNoLength-7))))
	// LUHN 算法处理
	return luhnProcess(preCardNo)
}

// luhnProcess 通过 LUHN 合成卡号处理给定的银行卡号
func luhnProcess(preCardNo string) string {
	checkSum := 0
	tmpCardNo := reverseString(preCardNo)
	for i, s := range tmpCardNo {
		// 数据层确保卡号正确
		tmp, _ := strconv.Atoi(string(s))
		// 由于卡号实际少了一位，所以反转后卡号第一位一定为偶数位
		// 同时 i 正好也是偶数，此时 i 将和卡号奇偶位同步
		if i%2 == 0 {
			// 偶数位 *2 是否为两位数(>9)
			if tmp*2 > 9 {
				// 如果为两位数则 -9
				checkSum += tmp*2 - 9
			} else {
				// 否则直接相加即可
				checkSum += tmp * 2
			}
		} else {
			// 奇数位直接相加
			checkSum += tmp
		}
	}
	if checkSum%10 != 0 {
		return preCardNo + strconv.Itoa(10-checkSum%10)
	} else {
		// 如果不巧生成的前 卡长度-1 位正好符合 LUHN 算法
		// 那么需要递归重新生成(需要符合 cardBind 中卡号长度)
		return RandBankNo()
	}
}

// RandEmail 返回随机邮箱，邮箱目前只支持常见的域名后缀
func RandEmail() string {
	return RandSmallLetters(8) + "@" + RandSmallLetters(5) + DomainSuffix[RandIntn(len(DomainSuffix))]
}

// RandIssueOrg 返回身份证签发机关(eg: XXX公安局/XX区分局)
func RandIssueOrg() string {
	return CityName[RandIntn(len(CityName))] + "公安局某某分局"
}

// RandValidPeriod 返回身份证有效期限(eg: 20150906-20350906)，有效期限固定为 20 年
func RandValidPeriod() string {
	begin := RandTime()
	end := begin.AddDate(20, 0, 0)
	return begin.Format("20060102") + "-" + end.Format("20060102")
}

// RandChinaID 返回中国大陆地区身份证号.
func RandChinaID() string {
	// AreaCode 随机一个+4位随机数字(不够左填充0)
	areaCode := AreaCode[RandIntn(len(AreaCode))] +
		fmt.Sprintf("%0*d", 4, RandIntBetween(1, 9999))
	birthday := RandTime().Format("20060102")
	randomCode := fmt.Sprintf("%0*d", 3, RandIntn(999))
	prefix := areaCode + birthday + randomCode
	return prefix + verifyCode(prefix)
}

// verifyCode 通过给定的身份证号生成最后一位的 verifyCode
func verifyCode(cardId string) string {
	tmp := 0
	for i, v := range Wi {
		t, _ := strconv.Atoi(string(cardId[i]))
		tmp += t * v
	}
	return ValCodeArr[tmp%11]
}

// RandMobile 返回中国大陆地区手机号
func RandMobile() string {
	return MobilePrefix[RandIntn(len(MobilePrefix))] + fmt.Sprintf("%0*d", 8, RandIntn(100000000))
}

// RandSex 返回性别
func RandSex() string {
	return If(RandBool(), "男", "女")
}

// RandChineseName 返回中国姓名，姓名已经尽量返回常用姓氏和名字
func RandChineseName() string {
	return Surnames[RandIntn(len(Surnames))] + RandChineseN(2)
}

// RandChineseN 指定长度随机中文字符(包含复杂字符)。
func RandChineseN(n int) string {
	var buf bytes.Buffer
	for i := 0; i < n; i++ {
		buf.WriteRune(rune(RandIntBetween(19968, 40869)))
	}
	return buf.String()
}

// RandChinese 指定范围随机中文字符.
func RandChinese(minLen, maxLen int) string {
	return RandChineseN(RandIntBetween(minLen, maxLen))
}

// RandSmallLetters 随机英文小写字母.
func RandSmallLetters(len int) string {
	data := make([]byte, len)
	for i := 0; i < len; i++ {
		data[i] = byte(RandIntn(26) + 97)
	}
	return string(data)
}

// 反转字符串
func reverseString(s string) string {
	runes := []rune(s)
	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}
	return string(runes)
}
