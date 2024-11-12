package ss

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
)

// ProvinceCity 返回随机省/城市
func (r random) ProvinceCity() string {
	return provinceCity[r.Intn(len(provinceCity))]
}

// Address 返回随机地址
func (r random) Address() string {
	return r.ProvinceCity() +
		r.Chinese(2, 3) + "路" +
		strconv.Itoa(r.IntBetween(1, 8000)) + "号" +
		r.Chinese(2, 3) + "小区" +
		strconv.Itoa(r.IntBetween(1, 20)) + "单元" +
		strconv.Itoa(r.IntBetween(101, 2500)) + "室"
}

// BankNo 返回随机银行卡号，银行卡号符合LUHN 算法并且有正确的卡 bin 前缀
func (r random) BankNo() string {
	// 随机选中银行卡卡头
	bank := cardBins[r.Intn(len(cardBins))]
	// 获取 卡前缀(cardBin)
	prefixes := bank.Prefixes
	// 获取当前银行卡正确长度
	cardNoLength := bank.Length
	// 生成 长度-1 位卡号
	preCardNo := strconv.Itoa(prefixes[r.Intn(len(prefixes))]) +
		fmt.Sprintf("%0*d", cardNoLength-7, r.Int64n(int64(math.Pow10(cardNoLength-7))))
	// LUHN 算法处理
	return r.luhnProcess(preCardNo)
}

// luhnProcess 通过 LUHN 合成卡号处理给定的银行卡号
func (r random) luhnProcess(preCardNo string) string {
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
		return r.BankNo()
	}
}

// Email 返回随机邮箱，邮箱目前只支持常见的域名后缀
func (r random) Email() string {
	return r.SmallLetters(8) + "@" + r.SmallLetters(5) + domainSuffix[r.Intn(len(domainSuffix))]
}

// IssueOrg 返回身份证签发机关(eg: XXX公安局/XX区分局)
func (r random) IssueOrg() string {
	return cityNames[r.Intn(len(cityNames))] + "公安局某某分局"
}

// ValidPeriod 返回身份证有效期限(eg: 20150906-20350906)，有效期限固定为 20 年
func (r random) ValidPeriod() string {
	begin := r.Time()
	end := begin.AddDate(20, 0, 0)
	return begin.Format("20060102") + "-" + end.Format("20060102")
}

// ChinaID 返回中国大陆地区身份证号.
func (r random) ChinaID() string {
	// AreaCode 随机一个+4位随机数字(不够左填充0)
	areaCode := AreaCode[r.Intn(len(AreaCode))] +
		fmt.Sprintf("%0*d", 4, r.IntBetween(1, 9999))
	birthday := r.Time().Format("20060102")
	randomCode := fmt.Sprintf("%0*d", 3, r.Intn(999))
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

// Mobile 返回中国大陆地区手机号
func (r random) Mobile() string {
	return mobilePrefix[r.Intn(len(mobilePrefix))] + fmt.Sprintf("%0*d", 8, r.Intn(100000000))
}

// Sex 返回性别
func (r random) Sex() string {
	return If(r.Bool(), "男", "女")
}

// ChineseName 返回中国姓名，姓名已经尽量返回常用姓氏和名字
func (r random) ChineseName() string {
	return Surnames[r.Intn(len(Surnames))] + r.ChineseN(2)
}

// ChineseN 指定长度随机中文字符(包含复杂字符)。
func (r random) ChineseN(n int) string {
	var buf bytes.Buffer
	for i := 0; i < n; i++ {
		buf.WriteRune(rune(r.IntBetween(19968, 40869)))
	}
	return buf.String()
}

// Chinese 指定范围随机中文字符.
func (r random) Chinese(minLen, maxLen int) string {
	return r.ChineseN(r.IntBetween(minLen, maxLen))
}

// SmallLetters 随机英文小写字母.
func (r random) SmallLetters(len int) string {
	data := make([]byte, len)
	for i := 0; i < len; i++ {
		data[i] = byte(r.Intn(26) + 97)
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
