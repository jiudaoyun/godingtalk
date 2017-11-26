package godingtalk

import (
    "fmt"
    "net/url"
	"gopkg.in/square/go-jose.v1/json"
	"strconv"
)

type Role struct {
	ID int
	Name string
	GroupName string
}

type User struct {
    OAPIResponse
    DingID string `json:"dingId"`
	UnionID string `json:"unionid"`
	OpenID string `json:"openId"`
    UserID string `json:"userid"`
    Name string

    Active bool

	StateCode string `json:"stateCode"` // 手机号国家码
    Mobile string
    Tel string
    IsHide bool // 是否号码隐藏, true表示隐藏, false表示不隐藏。隐藏手机号后，手机号在个人资料页隐藏，但仍可对其发DING、发起钉钉免费商务电话。

	Avatar string
    Workplace string
	Email string
	OrgEmail string
	Position string

    Remark string
    IsAdmin bool
    IsBoss bool
    IsSenior bool // 是否高管模式，true表示是，false表示不是。开启后，手机号码对所有员工隐藏。普通员工无法对其发DING、发起钉钉免费商务电话。高管之间不受影响。
	Departments []int `json:"department"`
	// TODO: parse, by self or otto?
    // IsLeaderInDepts map[int]string
	// OrderInDepts map[int]int
	IsLeaderInDepts string
	OrderInDepts string

    Extattr interface{}

    Roles []Role
}

type UserList struct {
    OAPIResponse
    HasMore bool
    Userlist []User
}

type Department struct {
    OAPIResponse
    Id int
    Name string
    ParentId int    
    Order int
    DeptPerimits string
    UserPerimits string
    OuterDept bool
    OuterPermitDepts string
    OuterPermitUsers string
    OrgDeptOwner string
    DeptManagerUseridList string
}

type DepartmentList struct {
    OAPIResponse
    Departments []Department `json:"department"`
}

type ExternalUser struct {
	UserID string `json:"userId"`
	Name string `json:"name"`
	Mobile string `json:"mobile"`
	Follower string `json:"follower_userid"` // 负责人userId
	Labels []int `json:"label_ids"` // 标签列表
	StateCode string `json:"state_code"` // 手机号国家码
	Company string `json:"company_name,omitempty"`
	Title string `json:"title,omitempty"`
	Email string `json:"email,omitempty"`
	Address string `json:"address,omitempty"`
	Remark string `json:"remark,omitempty"`
	SharedUsers []string `json:"share_userids,omitempty"` // 共享给的员工userId列表
	SharedDepts []int `json:"share_deptids,omitempty"` // 共享给的部门ID
}

type ExternalUserLabel struct {
	ID int
	Name string
}

type ExternalUserLabelGroup struct {
	Color int
	Name string
	Labels []ExternalUserLabel
}

// DepartmentList is 获取部门列表
func (c *DingTalkClient) DepartmentList() (DepartmentList, error) {
    var data DepartmentList
    err := c.httpRPC("department/list", nil, nil, &data)   
    return data, err
}

//DepartmentDetail is 获取部门详情
func (c *DingTalkClient) DepartmentDetail(id int) (*Department, error) {
    var data Department
    params := url.Values{}
    params.Add("id", fmt.Sprintf("%d", id))
    err :=c.httpRPC("department/get", params, nil, &data)
    return &data, err
}

//UserList is 获取部门成员
func (c *DingTalkClient) UserList(departmentID int) (UserList, error) {
    var data UserList
    params := url.Values{}
    params.Add("department_id", fmt.Sprintf("%d", departmentID))    
    err :=c.httpRPC("user/list", params, nil, &data)
    return data, err
}

//CreateChat is 
func (c *DingTalkClient) CreateChat(name string, owner string, useridlist []string) (string, error) {
    var data struct {
        OAPIResponse
        Chatid string
    }
    request := map[string]interface{} {
        "name":name,
        "owner":owner,
        "useridlist":useridlist,     
    }
    err :=c.httpRPC("chat/create", nil, request, &data)
    return data.Chatid, err
}

func (c *DingTalkClient) UserDetail(id string) (*User, error) {
	var user User
	params := url.Values{}
	params.Add("userid", id)
	err := c.httpRPC("user/get", params, nil, &user)
	return &user, err
}

type UserInfo struct {
	OAPIResponse
	UserID string `json:"userid"`
	DeviceID string `json:"deviceId"`
	IsSys bool `json:"is_sys"`
	SysLevel int `json:"sys_level"`
}

//UserInfoByCode 校验免登录码并换取用户身份
func (c *DingTalkClient) UserInfoByCode(code string) (*UserInfo, error) {
    var data UserInfo
    params := url.Values{}
    params.Add("code", code)
    err := c.httpRPC("user/getuserinfo", params, nil, &data)
    return &data, err
}

//UseridByUnionId 通过UnionId获取玩家Userid
func (c *DingTalkClient) UseridByUnionId(unionid string) (string, error) {
    var data struct {
		OAPIResponse
		UserID string `json:"userid"`
	}

    params := url.Values{}
    params.Add("unionid", unionid)
    err := c.httpRPC("user/getUseridByUnionid", params, nil, &data)
	if err!=nil {
		return "",err
	}

    return data.UserID, err
}

func (c *DingTalkClient) CreateExternalUser(euser *ExternalUser) (userID string, err error) {
	var rep struct {
		TaobaoOAPIResponse
		DingtalkCorpExtAddResponse struct{
			UserID string `json:"userid"`
		} `json:"dingtalk_corp_ext_add_response"`
	}

	d, _ := json.Marshal(euser)
	params := url.Values{}
	params.Add("contact", string(d))
	err = c.httpTaobaoRPC("dingtalk.corp.ext.add", params, &rep)
	if err != nil {
		return "", err
	}
	euser.UserID = rep.DingtalkCorpExtAddResponse.UserID
	return euser.UserID, nil
}

func (c *DingTalkClient) ExternalUserList(offset, size int) ([]ExternalUser, error) {
	type User struct {
		UserID string `json:"userId"`
		Name string `json:"name"`
		Mobile string `json:"mobile"`
		Follower string `json:"followerUserId"`
		Labels []int `json:"labelIds"`
		StateCode string `json:"stateCode"`
		Company string `json:"companyName"`
		Title string `json:"title"`
		Email string `json:"email"`
		Address string `json:"address"`
		Remark string `json:"remark"`
		SharedUsers []string `json:"shareUserIds"`
		SharedDepts []int `json:"shareDeptIds"`
	}
	var users []User

	var rep struct{
		TaobaoOAPIResponse
		Result string
	}

	params := url.Values{}
	params.Add("size", strconv.Itoa(size))
	params.Add("offset", strconv.Itoa(offset))
	err := c.httpTaobaoRPC("dingtalk.corp.ext.list", params, &rep)
	if err != nil {
		return nil, err
	}
	// fmt.Printf("result: %q\n", rep.Result)
	err = json.Unmarshal([]byte(rep.Result), &users)
	// fmt.Printf("users: %v\n", users)
	if err != nil {
		return nil, err
	}
	eusers := make([]ExternalUser, len(users))
	for i, user := range users {
		eusers[i] = ExternalUser{
			UserID: user.UserID,
			Name: user.Name,
			Mobile: user.Mobile,
			Follower: user.Follower,
			Labels: user.Labels,
			StateCode: user.StateCode,
			Company: user.Company,
			Title: user.Title,
			Email: user.Email,
			Address: user.Address,
			Remark: user.Remark,
			SharedUsers:user.SharedUsers,
			SharedDepts: user.SharedDepts,
		}
	}
	return eusers, nil
}

func (c *DingTalkClient) ExternalUserLabelGroups(offset, size int) ([]ExternalUserLabelGroup, error) {
	var rep struct{
		TaobaoOAPIResponse
		Result string
	}

	params := url.Values{}
	params.Add("size", strconv.Itoa(size))
	params.Add("offset", strconv.Itoa(offset))
	err := c.httpTaobaoRPC("dingtalk.corp.ext.listlabelgroups", params, &rep)
	if err != nil {
		return nil, err
	}
	fmt.Printf("labels: %s\n", rep.Result)
	var result []ExternalUserLabelGroup
	err = json.Unmarshal([]byte(rep.Result), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
