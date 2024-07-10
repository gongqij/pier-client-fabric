package main

type ladingBillInfo struct {
	LadingBillNumber          string `json:"ladingBillNumber"`          //1、提单编号
	BillType                  string `json:"billType"`                  //2、提单类型
	ShipperCode               string `json:"shipperCode"`               //3、托运人
	ShipperPhone              string `json:"shipperPhone"`              //4、托运人联系电话
	ShipperAddress            string `json:"shipperAddress"`            //5、托运人地址
	ConsigneeCode             string `json:"consigneeCode"`             //6、收货人
	ConsigneePhone            string `json:"consigneePhone"`            //7、收货人联系电话
	ConsigneeAddress          string `json:"consigneeAddress"`          //8、收货人地址
	NotifyPartyName           string `json:"notifyPartyName"`           //9、通知人
	NotifyPartyPhone          string `json:"notifyPartyPhone"`          //10、通知人电话
	NotifyPartyAddress        string `json:"notifyPartyAddress"`        //11、通知人地址
	PreCarriage               string `json:"preCarriage"`               //12、前段运输
	ReceiptPlace              string `json:"receiptPlace"`              //13、揽货地
	LoadingPort               string `json:"loadingPort"`               //14、装货港/站
	VoyNumber                 string `json:"voyNumber"`                 //15、航次/车次/车号
	DeliveryPort              string `json:"deliveryPort"`              //16、交付地
	DischargePort             string `json:"dischargePort"`             //17、卸货港/站
	Charges                   string `json:"charges"`                   //18、运费及费用说明
	CargoDeliveryAgency       string `json:"cargoDeliveryAgency"`       //19、提货地代理
	CompensationLimits        string `json:"compensationLimits"`        //20、计算赔偿限制总件数
	IssuePlace                string `json:"issuePlace"`                //21、签发地点
	IssueDate                 string `json:"issueDate"`                 //22、签发日期
	SignerStatement           string `json:"signerStatement"`           //23、签发人声明
	Mark                      string `json:"mark"`                      //24、唛头和号码
	ContainerNumber           string `json:"containerNumber"`           //25、箱号
	TotalReceivedNumber       string `json:"totalReceivedNumber"`       //26、集装箱/包装数
	GoodsDesc                 string `json:"goodsDesc"`                 //27、货物描述
	GrossWeight               string `json:"grossWeight"`               //28、毛重
	Size                      string `json:"size"`                      //29、尺码
	ContainerSealingNumber    string `json:"containerSealingNumber"`    //30、封号
	WaybillNumber             string `json:"waybillNumber"`             //31、运单号码
	ApplicantName             string `json:"applicantName"`             //32、申请人(企业)名称
	ApplicantSocialCreditCode string `json:"applicantSocialCreditCode"` //33、申请人(企业)社会信用代码
	HolderName                string `json:"holderName"`                //34、持有人(企业)名称
	HolderSocialCreditCode    string `json:"holderSocialCreditCode"`    //35、持有人(企业)社会信用代码
	IssueOrgName              string `json:"issueOrgName"`              //36、签发单位名称
	IssuerName                string `json:"issuerName"`                //37、签发人名称
	LadingBillStatus          string `json:"ladingBillStatus"`          //38、提单状态
}

type ladingBillCrossParams struct {
	CrossChainID           string         `json:"crossChainID"`           // 跨链唯一标识
	CorpHolderOrgCode      string         `json:"corpHolderOrgCode"`      // 持有企业组代号
	FinanceReceiverOrgCode string         `json:"financeReceiverOrgCode"` // 接收金融机构组代号
	LadingBillCR           ladingBillInfo `json:"ladingBillCR"`           // 提单信息
	Memo                   string         `json:"memo"`                   // 附言
}
