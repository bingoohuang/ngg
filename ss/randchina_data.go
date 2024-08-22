package ss

var AreaCode = [...]string{
	"11", "12", "13", "14", "15", "21", "22", "23", "31", "32", "33", "34", "35", "36", "37", "41", "42", "43", "44",
	"45", "46", "50", "51", "52", "53", "54", "61", "62", "63", "64", "65", "71", "81", "82", "91",
}

var AreaCodeMap = map[string]string{
	"11": "北京", "12": "天津", "13": "河北", "14": "山西", "15": "内蒙古", "21": "辽宁", "22": "吉林", "23": "黑龙江",
	"31": "上海", "32": "江苏", "33": "浙江", "34": "安徽", "35": "福建", "36": "江西", "37": "山东", "41": "河南", "42": "湖北",
	"43": "湖南", "44": "广东", "45": "广西", "46": "海南", "50": "重庆", "51": "四川", "52": "贵州", "53": "云南", "54": "西藏",
	"61": "陕西", "62": "甘肃", "63": "青海", "64": "宁夏", "65": "新疆", "71": "台湾", "81": "香港", "82": "澳门", "91": "国外",
}

var Surnames = []string{
	"赵", "钱", "孙", "李", "周", "吴", "郑", "王", "冯", "陈", "褚", "卫", "蒋", "沉", "韩", "杨", "朱", "秦", "尤", "许",
	"何", "吕", "施", "张", "孔", "曹", "严", "华", "金", "魏", "陶", "姜", "戚", "谢", "邹", "喻", "柏", "水", "窦", "章",
	"云", "苏", "潘", "葛", "奚", "范", "彭", "郎", "鲁", "韦", "昌", "马", "苗", "凤", "花", "方", "任", "袁", "柳", "鲍",
	"史", "唐", "费", "薛", "雷", "贺", "倪", "汤", "滕", "殷", "罗", "毕", "郝", "安", "常", "傅", "卞", "齐", "元", "顾",
	"孟", "平", "黄", "穆", "萧", "尹", "姚", "邵", "湛", "汪", "祁", "毛", "狄", "米", "伏", "成", "戴", "谈", "宋", "茅",
	"庞", "熊", "纪", "舒", "屈", "项", "祝", "董", "梁", "杜", "阮", "蓝", "闵", "季", "贾", "路", "娄", "江", "童", "颜",
	"郭", "梅", "盛", "林", "钟", "徐", "邱", "骆", "高", "夏", "蔡", "田", "樊", "胡", "凌", "霍", "虞", "万", "支", "柯",
	"管", "卢", "莫", "柯", "房", "裘", "缪", "解", "应", "宗", "丁", "宣", "邓", "单", "杭", "洪", "包", "诸", "左", "石",
	"崔", "吉", "龚", "程", "嵇", "邢", "裴", "陆", "荣", "翁", "荀", "于", "惠", "甄", "曲", "封", "储", "仲", "伊", "宁",
	"仇", "甘", "武", "符", "刘", "景", "詹", "龙", "叶", "幸", "司", "黎", "溥", "印", "怀", "蒲", "邰", "从", "索", "赖",
	"卓", "屠", "池", "乔", "胥", "闻", "莘", "党", "翟", "谭", "贡", "劳", "逄", "姬", "申", "扶", "堵", "冉", "宰", "雍",
	"桑", "寿", "通", "燕", "浦", "尚", "农", "温", "别", "庄", "晏", "柴", "瞿", "阎", "连", "习", "容", "向", "古", "易",
	"廖", "庾", "终", "步", "都", "耿", "满", "弘", "匡", "国", "文", "寇", "广", "禄", "阙", "东", "欧", "利", "师", "巩",
	"聂", "关", "荆", "司马", "上官", "欧阳", "夏侯", "诸葛", "闻人", "东方", "赫连", "皇甫", "尉迟", "公羊", "澹台", "公冶",
	"宗政", "濮阳", "淳于", "单于", "太叔", "申屠", "公孙", "仲孙", "轩辕", "令狐", "徐离", "宇文", "长孙", "慕容", "司徒", "司空",
}

var CityName = [...]string{
	"北京市", "上海市", "天津市", "重庆市", "石家庄市", "唐山市", "秦皇岛市", "邯郸市", "邢台市", "保定市", "张家口市", "承德市",
	"沧州市", "廊坊市", "衡水市", "太原市", "大同市", "阳泉市", "长治市", "晋城市", "朔州市", "晋中市", "运城市", "忻州市", "临汾市",
	"吕梁市", "呼和浩特市", "包头市", "乌海市", "赤峰市", "通辽市", "鄂尔多斯市", "呼伦贝尔市", "巴彦淖尔市", "乌兰察布市", "兴安盟", "锡林郭勒盟",
	"阿拉善盟", "沈阳市", "大连市", "鞍山市", "抚顺市", "本溪市", "丹东市", "锦州市", "营口市", "阜新市", "辽阳市", "盘锦市",
	"铁岭市", "朝阳市", "葫芦岛市", "长春市", "吉林市", "四平市", "辽源市", "通化市", "白山市", "松原市", "白城市", "延边朝鲜族自治州", "哈尔滨市",
	"齐齐哈尔市", "鸡西市", "鹤岗市", "双鸭山市", "大庆市", "伊春市", "佳木斯市", "七台河市", "牡丹江市", "黑河市", "绥化市", "大兴安岭地区",
	"南京市", "无锡市", "徐州市", "常州市", "苏州市", "南通市", "连云港市", "淮安市", "盐城市", "扬州市", "镇江市", "泰州市", "宿迁市", "杭州市", "宁波市",
	"温州市", "嘉兴市", "湖州市", "绍兴市", "金华市", "衢州市", "舟山市", "台州市", "丽水市", "合肥市", "芜湖市", "蚌埠市", "淮南市", "马鞍山市", "淮北市", "铜陵市",
	"安庆市", "黄山市", "滁州市", "阜阳市", "宿州市", "六安市", "亳州市", "池州市", "宣城市", "福州市", "厦门市", "莆田市", "三明市", "泉州市", "漳州市", "南平市",
	"龙岩市", "宁德市", "南昌市", "景德镇市", "萍乡市", "九江市", "新余市", "鹰潭市", "赣州市", "吉安市", "宜春市", "抚州市", "上饶市", "济南市", "青岛市", "淄博市",
	"枣庄市", "东营市", "烟台市", "潍坊市", "济宁市", "泰安市", "威海市", "日照市", "莱芜市", "临沂市", "德州市", "聊城市", "滨州市", "菏泽市", "郑州市", "开封市",
	"洛阳市", "平顶山市", "安阳市", "鹤壁市", "新乡市", "焦作市", "濮阳市", "许昌市", "漯河市", "三门峡市", "南阳市", "商丘市", "信阳市", "周口市", "驻马店市",
	"省直辖县级行政区划", "武汉市", "黄石市", "十堰市", "宜昌市", "襄阳市", "鄂州市", "荆门市", "孝感市", "荆州市", "黄冈市", "咸宁市", "随州市", "恩施土家族苗族自治州",
	"省直辖县级行政区划", "长沙市", "株洲市", "湘潭市", "衡阳市", "邵阳市", "岳阳市", "常德市", "张家界市", "益阳市", "郴州市", "永州市", "怀化市", "娄底市",
	"湘西土家族苗族自治州", "广州市", "韶关市", "深圳市", "珠海市", "汕头市", "佛山市", "江门市", "湛江市", "茂名市", "肇庆市", "惠州市", "梅州市", "汕尾市",
	"河源市", "阳江市", "清远市", "东莞市", "中山市", "潮州市", "揭阳市", "云浮市", "南宁市", "柳州市", "桂林市", "梧州市", "北海市", "防城港市", "钦州市", "贵港市",
	"玉林市", "百色市", "贺州市", "河池市", "来宾市", "崇左市", "海口市", "三亚市", "三沙市", "成都市", "自贡市", "攀枝花市", "泸州市", "德阳市", "绵阳市", "广元市",
	"遂宁市", "内江市", "乐山市", "南充市", "眉山市", "宜宾市", "广安市", "达州市", "雅安市", "巴中市", "资阳市", "阿坝藏族羌族自治州", "甘孜藏族自治州", "凉山彝族自治州",
	"贵阳市", "六盘水市", "遵义市", "安顺市", "毕节市", "铜仁市", "黔西南布依族苗族自治州", "黔东南苗族侗族自治州", "黔南布依族苗族自治州", "昆明市", "曲靖市", "玉溪市",
	"保山市", "昭通市", "丽江市", "普洱市", "临沧市", "楚雄彝族自治州", "红河哈尼族彝族自治州", "文山壮族苗族自治州", "西双版纳傣族自治州", "大理白族自治州",
	"德宏傣族景颇族自治州", "怒江傈僳族自治州", "迪庆藏族自治州", "拉萨市", "日喀则市", "昌都地区", "山南地区", "那曲地区", "阿里地区", "林芝地区", "西安市", "铜川市",
	"宝鸡市", "咸阳市", "渭南市", "延安市", "汉中市", "榆林市", "安康市", "商洛市", "兰州市", "嘉峪关市", "金昌市", "白银市", "天水市", "武威市", "张掖市", "平凉市",
	"酒泉市", "庆阳市", "定西市", "陇南市", "临夏回族自治州", "甘南藏族自治州", "西宁市", "海东市", "海北藏族自治州", "黄南藏族自治州", "海南藏族自治州", "果洛藏族自治州",
	"玉树藏族自治州", "海西蒙古族藏族自治州", "银川市", "石嘴山市", "吴忠市", "固原市", "中卫市", "乌鲁木齐市", "克拉玛依市", "吐鲁番地区", "哈密地区", "昌吉回族自治州",
	"博尔塔拉蒙古自治州", "巴音郭楞蒙古自治州", "阿克苏地区", "克孜勒苏柯尔克孜自治州", "喀什地区", "和田地区", "伊犁哈萨克自治州", "塔城地区", "阿勒泰地区",
}

var DomainSuffix = [...]string{
	".biz", ".cloud", ".club", ".cn", ".co", ".com", ".com.cn", ".info", ".me", ".net", ".org", ".space", ".store",
	".us", ".vip", ".xyz",
}

var MobilePrefix = [...]string{
	"130", "131", "132", "133", "134", "135", "136", "137", "138", "139", "145", "147", "150", "151", "152",
	"153", "155", "156", "157", "158", "159", "170", "176", "177", "178", "180", "181", "182", "183", "184",
	"185", "186", "187", "188", "189",
}

var ValCodeArr = [...]string{
	"1", "0", "X", "9", "8", "7", "6", "5", "4", "3", "2",
}

var Wi = [...]int{
	7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2,
}

var ProvinceCity = [...]string{
	"黑龙江省齐齐哈尔市",
	"黑龙江省黑河市",
	"黑龙江省鹤岗市",
	"黑龙江省鸡西市",
	"黑龙江省绥化市",
	"黑龙江省牡丹江市",
	"黑龙江省大庆市",
	"黑龙江省大兴安岭地区",
	"黑龙江省哈尔滨市",
	"黑龙江省双鸭山市",
	"黑龙江省佳木斯市",
	"黑龙江省伊春市",
	"黑龙江省七台河市",
	"香港特别行政区香港特别行政区",
	"青海省黄南藏族自治州",
	"青海省西宁市",
	"青海省玉树藏族自治州",
	"青海省海西蒙古族藏族自治州",
	"青海省海南藏族自治州",
	"青海省海北藏族自治州",
	"青海省海东地区",
	"青海省果洛藏族自治州",
	"陕西省铜川市",
	"陕西省西安市",
	"陕西省渭南市",
	"陕西省汉中市",
	"陕西省榆林市",
	"陕西省延安市",
	"陕西省宝鸡市",
	"陕西省安康市",
	"陕西省商洛市",
	"陕西省咸阳市",
	"重庆市重庆市",
	"辽宁省鞍山市",
	"辽宁省阜新市",
	"辽宁省锦州市",
	"辽宁省铁岭市",
	"辽宁省辽阳市",
	"辽宁省葫芦岛市",
	"辽宁省营口市",
	"辽宁省盘锦市",
	"辽宁省沈阳市",
	"辽宁省本溪市",
	"辽宁省朝阳市",
	"辽宁省抚顺市",
	"辽宁省大连市",
	"辽宁省丹东市",
	"贵州省黔西南布依族苗族自治州",
	"贵州省黔南布依族苗族自治州",
	"贵州省黔东南苗族侗族自治州",
	"贵州省铜仁地区",
	"贵州省遵义市",
	"贵州省贵阳市",
	"贵州省毕节地区",
	"贵州省安顺市",
	"贵州省六盘水市",
	"西藏自治区阿里地区",
	"西藏自治区那曲地区",
	"西藏自治区林芝地区",
	"西藏自治区昌都地区",
	"西藏自治区日喀则地区",
	"西藏自治区拉萨市",
	"西藏自治区山南地区",
	"福建省龙岩市",
	"福建省莆田市",
	"福建省福州市",
	"福建省漳州市",
	"福建省泉州市",
	"福建省宁德市",
	"福建省厦门市",
	"福建省南平市",
	"福建省三明市",
	"甘肃省陇南市",
	"甘肃省金昌市",
	"甘肃省酒泉市",
	"甘肃省白银市",
	"甘肃省甘南藏族自治州",
	"甘肃省武威市",
	"甘肃省张掖市",
	"甘肃省庆阳市",
	"甘肃省平凉市",
	"甘肃省定西市",
	"甘肃省天水市",
	"甘肃省嘉峪关市",
	"甘肃省兰州市",
	"甘肃省临夏回族自治州",
	"澳门特别行政区澳门特别行政区",
	"湖南省长沙市",
	"湖南省郴州市",
	"湖南省邵阳市",
	"湖南省衡阳市",
	"湖南省益阳市",
	"湖南省湘西土家族苗族自治州",
	"湖南省湘潭市",
	"湖南省永州市",
	"湖南省株洲市",
	"湖南省怀化市",
	"湖南省张家界市",
	"湖南省常德市",
	"湖南省岳阳市",
	"湖南省娄底市",
	"湖北省黄石市",
	"湖北省黄冈市",
	"湖北省随州市",
	"湖北省鄂州市",
	"湖北省襄樊市",
	"湖北省荆门市",
	"湖北省荆州市",
	"湖北省神农架",
	"湖北省武汉市",
	"湖北省恩施土家族苗族自治州",
	"湖北省宜昌市",
	"湖北省孝感市",
	"湖北省咸宁市",
	"湖北省十堰市",
	"海南省海口市",
	"海南省三亚市",
	"浙江省金华市",
	"浙江省衢州市",
	"浙江省舟山市",
	"浙江省绍兴市",
	"浙江省湖州市",
	"浙江省温州市",
	"浙江省杭州市",
	"浙江省宁波市",
	"浙江省嘉兴市",
	"浙江省台州市",
	"浙江省丽水市",
	"河南省鹤壁市",
	"河南省驻马店市",
	"河南省郑州市",
	"河南省许昌市",
	"河南省焦作市",
	"河南省濮阳市",
	"河南省漯河市",
	"河南省洛阳市",
	"河南省新乡市",
	"河南省开封市",
	"河南省平顶山市",
	"河南省安阳市",
	"河南省商丘市",
	"河南省周口市",
	"河南省南阳市",
	"河南省信阳市",
	"河南省三门峡市",
	"河北省邯郸市",
	"河北省邢台市",
	"河北省衡水市",
	"河北省秦皇岛市",
	"河北省石家庄市",
	"河北省沧州市",
	"河北省承德市",
	"河北省张家口市",
	"河北省廊坊市",
	"河北省唐山市",
	"河北省保定市",
	"江西省鹰潭市",
	"江西省赣州市",
	"江西省萍乡市",
	"江西省景德镇市",
	"江西省新余市",
	"江西省抚州市",
	"江西省宜春市",
	"江西省吉安市",
	"江西省南昌市",
	"江西省九江市",
	"江西省上饶市",
	"江苏省镇江市",
	"江苏省连云港市",
	"江苏省苏州市",
	"江苏省盐城市",
	"江苏省淮安市",
	"江苏省泰州市",
	"江苏省无锡市",
	"江苏省扬州市",
	"江苏省徐州市",
	"江苏省常州市",
	"江苏省宿迁市",
	"江苏省南通市",
	"江苏省南京市",
	"新疆维吾尔自治区阿拉尔市",
	"新疆维吾尔自治区阿勒泰地区",
	"新疆维吾尔自治区阿克苏地区",
	"新疆维吾尔自治区石河子市",
	"新疆维吾尔自治区昌吉回族自治州",
	"新疆维吾尔自治区巴音郭楞蒙古自治州",
	"新疆维吾尔自治区塔城地区",
	"新疆维吾尔自治区图木舒克市",
	"新疆维吾尔自治区喀什地区",
	"新疆维吾尔自治区哈密地区",
	"新疆维吾尔自治区和田地区",
	"新疆维吾尔自治区吐鲁番地区",
	"新疆维吾尔自治区博尔塔拉蒙古自治州",
	"新疆维吾尔自治区克拉玛依市",
	"新疆维吾尔自治区克孜勒苏柯尔克孜自治州",
	"新疆维吾尔自治区伊犁哈萨克自治州",
	"新疆维吾尔自治区五家渠市",
	"新疆维吾尔自治区乌鲁木齐市",
	"广西壮族自治区防城港市",
	"广西壮族自治区钦州市",
	"广西壮族自治区贺州市",
	"广西壮族自治区贵港市",
	"广西壮族自治区百色市",
	"广西壮族自治区玉林市",
	"广西壮族自治区河池市",
	"广西壮族自治区梧州市",
	"广西壮族自治区桂林市",
	"广西壮族自治区柳州市",
	"广西壮族自治区来宾市",
	"广西壮族自治区崇左市",
	"广西壮族自治区南宁市",
	"广西壮族自治区北海市",
	"广东省韶关市",
	"广东省阳江市",
	"广东省茂名市",
	"广东省肇庆市",
	"广东省珠海市",
	"广东省潮州市",
	"广东省湛江市",
	"广东省清远市",
	"广东省深圳市",
	"广东省河源市",
	"广东省江门市",
	"广东省汕尾市",
	"广东省汕头市",
	"广东省梅州市",
	"广东省揭阳市",
	"广东省惠州市",
	"广东省广州市",
	"广东省佛山市",
	"广东省云浮市",
	"广东省中山市",
	"广东省东莞市",
	"山西省阳泉市",
	"山西省长治市",
	"山西省运城市",
	"山西省朔州市",
	"山西省晋城市",
	"山西省晋中市",
	"山西省忻州市",
	"山西省太原市",
	"山西省大同市",
	"山西省吕梁市",
	"山西省临汾市",
	"山东省青岛市",
	"山东省菏泽市",
	"山东省莱芜市",
	"山东省聊城市",
	"山东省烟台市",
	"山东省潍坊市",
	"山东省滨州市",
	"山东省淄博市",
	"山东省济宁市",
	"山东省济南市",
	"山东省泰安市",
	"山东省枣庄市",
	"山东省日照市",
	"山东省德州市",
	"山东省威海市",
	"山东省临沂市",
	"山东省东营市",
	"安徽省黄山市",
	"安徽省马鞍山市",
	"安徽省阜阳市",
	"安徽省铜陵市",
	"安徽省蚌埠市",
	"安徽省芜湖市",
	"安徽省滁州市",
	"安徽省淮南市",
	"安徽省淮北市",
	"安徽省池州市",
	"安徽省巢湖市",
	"安徽省宿州市",
	"安徽省宣城市",
	"安徽省安庆市",
	"安徽省合肥市",
	"安徽省六安市",
	"安徽省亳州市",
	"宁夏回族自治区银川市",
	"宁夏回族自治区石嘴山市",
	"宁夏回族自治区固原市",
	"宁夏回族自治区吴忠市",
	"宁夏回族自治区中卫市",
	"天津市天津市",
	"四川省雅安市",
	"四川省阿坝藏族羌族自治州",
	"四川省遂宁市",
	"四川省达州市",
	"四川省资阳市",
	"四川省自贡市",
	"四川省绵阳市",
	"四川省眉山市",
	"四川省甘孜藏族自治州",
	"四川省泸州市",
	"四川省攀枝花市",
	"四川省成都市",
	"四川省德阳市",
	"四川省广安市",
	"四川省广元市",
	"四川省巴中市",
	"四川省宜宾市",
	"四川省南充市",
	"四川省凉山彝族自治州",
	"四川省内江市",
	"四川省乐山市",
	"吉林省长春市",
	"吉林省通化市",
	"吉林省辽源市",
	"吉林省白山市",
	"吉林省白城市",
	"吉林省松原市",
	"吉林省延边朝鲜族自治州",
	"吉林省四平市",
	"吉林省吉林市",
	"台湾省台湾省",
	"北京市北京市",
	"内蒙古自治区阿拉善盟",
	"内蒙古自治区锡林郭勒盟",
	"内蒙古自治区鄂尔多斯市",
	"内蒙古自治区通辽市",
	"内蒙古自治区赤峰市",
	"内蒙古自治区巴彦淖尔市",
	"内蒙古自治区呼和浩特市",
	"内蒙古自治区呼伦贝尔市",
	"内蒙古自治区包头市",
	"内蒙古自治区兴安盟",
	"内蒙古自治区乌海市",
	"内蒙古自治区乌兰察布市",
	"云南省迪庆藏族自治州",
	"云南省西双版纳傣族自治州",
	"云南省红河哈尼族彝族自治州",
	"云南省玉溪市",
	"云南省楚雄彝族自治州",
	"云南省曲靖市",
	"云南省普洱市",
	"云南省昭通市",
	"云南省昆明市",
	"云南省文山壮族苗族自治州",
	"云南省怒江傈僳族自治州",
	"云南省德宏傣族景颇族自治州",
	"云南省大理白族自治州",
	"云南省保山市",
	"云南省丽江市",
	"云南省临沧市",
	"上海市上海市",
}

type CardBin struct {
	Name     string
	Length   int
	CardType string
	Prefixes []int
}

var CardBins = [...]CardBin{
	{
		"工商银行",
		19,
		"借记卡",
		[]int{
			620058, 621225, 621226, 621227, 621281, 621288, 621558, 621559, 621670, 621721, 621722, 621723, 622200, 622202, 622203, 622208, 622307, 622902, 623062,
		},
	},
	{
		"农业银行",
		19,
		"借记卡",
		[]int{
			621282, 621336, 621671, 622821, 622822, 622823, 622825, 622827, 622828, 622841, 622843, 622845, 622846, 622848, 622849,
		},
	},
	{
		"中国银行",
		19,
		"借记卡",
		[]int{
			456351, 601382, 620061, 621283, 621330, 621332, 621568, 621569, 621660, 621661, 621663, 621666, 621668, 621669, 621672, 621725, 621756, 621758, 621785, 621786, 621787, 621788, 621790, 623208, 623569, 623571, 623572, 623573, 623575, 623586,
		},
	},
	{
		"建设银行",
		16,
		"借记卡",
		[]int{
			421349, 434061, 434062, 436742, 524094, 526410, 552245, 620060, 621080, 621081, 621082, 621284, 621466, 621467, 621488, 621499, 621598, 621673, 621700, 622280, 622700, 622966, 622988, 623094, 623211, 623668,
		},
	},
	{
		"兴业银行",
		18,
		"借记卡",
		[]int{
			622908, 622909, 966666,
		},
	},
	{
		"光大银行",
		16,
		"借记卡",
		[]int{
			620518, 621489, 621491, 621492, 622661, 622662, 622663, 622664, 622665, 622666, 622667, 622668, 622669, 622670, 622673, 623156,
		},
	},
	{
		"中信银行",
		16,
		"借记卡",
		[]int{
			442729, 442730, 621768, 621771, 621773, 622690, 622691, 622696, 622698, 622998, 968807,
		},
	},
	{
		"平安银行",
		16,
		"借记卡",
		[]int{
			602907, 621626, 622298, 622538, 622986, 622989, 623058, 627066,
		},
	},
	{
		"民生银行",
		16,
		"借记卡",
		[]int{
			415599, 421393, 421865, 427570, 472067, 472068, 621691, 622616, 622617, 622618, 622619, 622620, 622622, 623683,
		},
	},
	{
		"广发银行股份有限公司",
		19,
		"借记卡",
		[]int{
			621462, 622568,
		},
	},
	{
		"浦东发展银行",
		16,
		"借记卡",
		[]int{
			621351, 621792, 621793, 621795, 622516, 622518, 622521, 622522, 622523,
		},
	},
	{
		"交通银行",
		17,
		"借记卡",
		[]int{
			405512, 601428, 622258, 622260, 622262,
		},
	},
	{
		"邮储银行",
		19,
		"借记卡",
		[]int{
			620062, 621095, 621098, 621285, 621599, 621797, 621798, 621799, 622150, 622151, 622188, 623218, 623698, 955100,
		},
	},
	{
		"招商银行",
		16,
		"借记卡",
		[]int{
			410062, 468203, 512425, 524011, 621286, 621483, 621485, 621486, 622580, 622588, 622609,
		},
	},
	{
		"北京农村商业银行",
		19,
		"借记卡",
		[]int{
			621067,
		},
	},
	{
		"渤海银行",
		16,
		"借记卡",
		[]int{
			621453, 622884,
		},
	},
	{
		"常熟农村商业银行",
		19,
		"借记卡",
		[]int{
			622323,
		},
	},
	{
		"长安银行",
		19,
		"借记卡",
		[]int{
			621448,
		},
	},
	{
		"德阳银行",
		19,
		"借记卡",
		[]int{
			622561,
		},
	},
	{
		"福建海峡银行股份有限公司",
		18,
		"借记卡",
		[]int{
			621267,
		},
	},
	{
		"福建省农村信用社联合社",
		19,
		"借记卡",
		[]int{
			622184, 623036,
		},
	},
	{
		"广东省农村信用社联合社",
		19,
		"借记卡",
		[]int{
			621518, 621728,
		},
	},
	{
		"广东顺德农村商业银行",
		16,
		"借记卡",
		[]int{
			622322,
		},
	},
	{
		"广州农村商业银行股份有限公司",
		18,
		"借记卡",
		[]int{
			622439,
		},
	},
	{
		"广州银行股份有限公司",
		19,
		"借记卡",
		[]int{
			622467,
		},
	},
	{
		"桂林市商业银行",
		17,
		"借记卡",
		[]int{
			622856,
		},
	},
	{
		"哈尔滨银行",
		17,
		"借记卡",
		[]int{
			622425,
		},
	},
	{
		"邯郸银行",
		18,
		"借记卡",
		[]int{
			622960,
		},
	},
	{
		"河北银行股份有限公司",
		19,
		"借记卡",
		[]int{
			623000,
		},
	},
	{
		"湖北农信社",
		16,
		"借记卡",
		[]int{
			621013, 622412,
		},
	},
	{
		"湖南省农村信用社联合社",
		19,
		"借记卡",
		[]int{
			622169, 623090,
		},
	},
	{
		"黄河农村商业银行",
		19,
		"借记卡",
		[]int{
			622947, 623095,
		},
	},
	{
		"吉林农信联合社",
		19,
		"借记卡",
		[]int{
			622935,
		},
	},
	{
		"江苏农信社",
		19,
		"借记卡",
		[]int{
			622324,
		},
	},
	{
		"江苏省农村信用社联合社",
		19,
		"借记卡",
		[]int{
			623066,
		},
	},
	{
		"江苏银行",
		19,
		"借记卡",
		[]int{
			622173, 622873, 622876,
		},
	},
	{
		"江西农信联合社",
		19,
		"借记卡",
		[]int{
			622682,
		},
	},
	{
		"江西银行",
		16,
		"借记卡",
		[]int{
			621269, 621359, 622275,
		},
	},
	{
		"九江银行股份有限公司",
		19,
		"借记卡",
		[]int{
			622162, 623146,
		},
	},
	{
		"昆明农联社",
		16,
		"借记卡",
		[]int{
			622369, 623190,
		},
	},
	{
		"龙江银行",
		16,
		"借记卡",
		[]int{
			622860,
		},
	},
	{
		"南充市商业银行",
		19,
		"借记卡",
		[]int{
			623072,
		},
	},
	{
		"南京银行",
		16,
		"借记卡",
		[]int{
			621777,
		},
	},
	{
		"内蒙古自治区农村信用联合社",
		19,
		"借记卡",
		[]int{
			621737, 622976,
		},
	},
	{
		"宁波银行",
		19,
		"借记卡",
		[]int{
			621418, 622281, 622316,
		},
	},
	{
		"齐鲁银行股份有限公司",
		19,
		"借记卡",
		[]int{
			622379,
		},
	},
	{
		"秦皇岛银行股份有限公司",
		19,
		"借记卡",
		[]int{
			621237,
		},
	},
	{
		"青岛银行",
		19,
		"借记卡",
		[]int{
			623170,
		},
	},
	{
		"青海省农村信用社联合社",
		16,
		"借记卡",
		[]int{
			621517,
		},
	},
	{
		"山东省农村信用社联合社",
		16,
		"借记卡",
		[]int{
			621521, 622319, 622320,
		},
	},
	{
		"上海农商银行",
		19,
		"借记卡",
		[]int{
			623162,
		},
	},
	{
		"上海银行",
		18,
		"借记卡",
		[]int{
			620522, 622267, 622279, 622468, 622892,
		},
	},
	{
		"深圳农村商业银行",
		16,
		"借记卡",
		[]int{
			623035,
		},
	},
	{
		"台州银行",
		19,
		"借记卡",
		[]int{
			623039,
		},
	},
	{
		"泰安银行",
		19,
		"借记卡",
		[]int{
			623196,
		},
	},
	{
		"温州银行",
		16,
		"借记卡",
		[]int{
			621977,
		},
	},
	{
		"浙江稠州商业银行",
		16,
		"借记卡",
		[]int{
			621028,
		},
	},
	{
		"浙江民泰商业银行",
		19,
		"借记卡",
		[]int{
			621088, 621726,
		},
	},
}
