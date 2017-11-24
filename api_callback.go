package godingtalk

const (
	EVENT_USER_ADD_ORG = "user_add_org"  // 通讯录用户增加
	EVENT_USER_MODIFY_ORG = "user_modify_org"  // 通讯录用户更改
	EVENT_USER_LEAVE_ORG = "user_leave_org"  // 通讯录用户离职
	EVENT_ORG_ADMIN_ADD = "org_admin_add"  // 通讯录用户被设为管理员
	EVENT_ORG_ADMIN_REMOVE = "org_admin_remove"  // 通讯录用户被取消设置管理员
	EVENT_ORG_DEPT_CREATE = "org_dept_create"  // 通讯录企业部门创建
	EVENT_ORG_DEPT_MODIFY = "org_dept_modify"  // 通讯录企业部门修改
	EVENT_ORG_DEPT_REMOVE = "org_dept_remove"  // 通讯录企业部门删除
	EVENT_ORG_REMOVE = "org_remove"  // 企业被解散
	EVENT_ORG_CHANGE = "org_change"  // 企业信息发生变更
	EVENT_LABEL_USER_CHANGE = "label_user_change"  // 员工角色信息发生变更
	EVENT_LABEL_CONF_ADD = "label_conf_add"  // 增加角色或者角色组
	EVENT_LABEL_CONF_DEL = "label_conf_del"  // 删除角色或者角色组
	EVENT_LABEL_CONF_MODIFY = "label_conf_modify"  // 修改角色或者角色组
)

type ContactEvent struct {
	EventType string `json:"EventType"`
	TimeStamp int64 `json:"TimeStamp"`
	UserIDs []string `json:"UserId"`
	DeptIDs []string `json:"DeptId"`
	CorpID string `json:"CorpId"`
}

type Callback struct {
	OAPIResponse
	Token     string
	AES_KEY   string `json:"aes_key"`
	URL       string
	Callbacks []string `json:"call_back_tag"`
}

//RegisterCallback is 注册事件回调接口
func (c *DingTalkClient) RegisterCallback(callbacks []string, token string, aes_key string, callbackURL string) error {
	var data OAPIResponse
	request := map[string]interface{}{
		"call_back_tag": callbacks,
		"token":         token,
		"aes_key":       aes_key,
		"url":           callbackURL,
	}
	err := c.httpRPC("call_back/register_call_back", nil, request, &data)
	return err
}

//UpdateCallback is 更新事件回调接口
func (c *DingTalkClient) UpdateCallback(callbacks []string, token string, aes_key string, callbackURL string) error {
	var data OAPIResponse
	request := map[string]interface{}{
		"call_back_tag": callbacks,
		"token":         token,
		"aes_key":       aes_key,
		"url":           callbackURL,
	}
	err := c.httpRPC("call_back/update_call_back", nil, request, &data)
	return err
}

//DeleteCallback is 删除事件回调接口
func (c *DingTalkClient) DeleteCallback() error {
	var data OAPIResponse
	err := c.httpRPC("call_back/delete_call_back", nil, nil, &data)
	return err
}

//ListCallback is 查询事件回调接口
func (c *DingTalkClient) ListCallback() (Callback, error) {
	var data Callback
	err := c.httpRPC("call_back/get_call_back", nil, nil, &data)
	return data, err
}
