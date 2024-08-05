package main

type ladingBillInfo struct {
	Tdbh      string `json:"tdbh"`      // 1、提单编号
	Tltdlx    string `json:"tltdlx"`    // 2、提单类型
	Tyrmc     string `json:"tyrmc"`     // 3、托运人
	Tyrdh     string `json:"tyrdh"`     // 4、托运人联系电话
	Tyrdz     string `json:"tyrdz"`     // 5、托运人地址
	Shrhpzs   string `json:"shrhpzs"`   // 6、收货人
	Shrlxdh   string `json:"shrlxdh"`   // 7、收货人联系电话
	Shrdz     string `json:"shrdz"`     // 8、收货人地址
	Tzrmc     string `json:"tzrmc"`     // 9、通知人
	Tzrdh     string `json:"tzrdh"`     // 10、通知人电话
	Tzrdz     string `json:"tzrdz"`     // 11、通知人地址
	Qdys      string `json:"qdys"`      // 12、前段运输
	Shd       string `json:"shd"`       // 13、揽货地
	Zhg       string `json:"zhg"`       // 14、装货港/站
	Cc        string `json:"cc"`        // 15、航次/车次/车号
	Jfd       string `json:"jfd"`       // 16、交付地
	Xhg       string `json:"xhg"`       // 17、卸货港/站
	Fy        string `json:"fy"`        // 18、运费及费用说明
	Thddl     string `json:"thddl"`     // 19、提货地代理
	Zjs       string `json:"zjs"`       // 20、计算赔偿限制总件数
	Qfdd      string `json:"qfdd"`      // 21、签发地点
	Qfrq      string `json:"qfrq"`      // 22、签发日期
	Qfrsm     string `json:"qfrsm"`     // 23、签发人声明
	Mthhm     string `json:"mthhm"`     // 24、唛头和号码
	Xh        string `json:"xh"`        // 25、箱号
	Jzx       string `json:"jzx"`       // 26、集装箱/包装数
	Hwms      string `json:"hwms"`      // 27、货物描述
	Mz        string `json:"mz"`        // 28、毛重
	Cm        string `json:"cm"`        // 29、尺码
	Fh        string `json:"fh"`        // 30、封号
	Ydhm      string `json:"ydhm"`      // 31、运单号码
	Sqrmc     string `json:"sqrmc"`     // 32、申请人(企业)名称
	Sqrshxydm string `json:"sqrshxydm"` // 33、申请人(企业)社会信用代码
	Tdcyrmc   string `json:"tdcyrmc"`   // 34、持有人(企业)名称
	Cyrshxydm string `json:"cyrshxydm"` // 35、持有人(企业)社会信用代码
	Qfdwmc    string `json:"qfdwmc"`    // 36、签发单位名称
	Qfrmc     string `json:"qfrmc"`     // 37、签发人名称
	Tdzt      string `json:"tdzt"`      // 38、提单状态
}

type ladingBillCrossParams struct {
	LadingBillNumber       string         `json:"ladingBillNumber"`       // 跨链唯一标识
	CorpHolderOrgCode      string         `json:"corpHolderOrgCode"`      // 持有企业组代号
	FinanceReceiverOrgCode string         `json:"financeReceiverOrgCode"` // 接收金融机构组代号
	LadingBillCR           ladingBillInfo `json:"ladingBillCR"`           // 提单信息
	Memo                   string         `json:"memo"`                   // 附言
}
