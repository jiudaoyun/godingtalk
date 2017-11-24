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
    UserID string `json:"userid"`
    Name string

    Active bool

    Mobile string
    Tel string
    IsHide bool

    Workplace string
	Email string
	OrgEmail string

    Remark string
    Order int
    IsAdmin bool
    IsBoss bool
	Departments []int `json:"department"`
    IsLeaderInDepts map[int]string
	OrderInDepts map[int]int
	IsSys bool `json:"is_sys"`
	SysLevel int `json:"sys_level"`
    Position string
    Avatar string

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
	Name string
	Mobile string
	Follower string `json:"follower_userid"` // 负责人userId
	Labels []int `json:"label_ids"` // 标签列表
	StateCode string `json:"state_code"` // 手机号国家码
	Company string `json:"company_name,omitempty"`
	Title string `json:",omitempty"`
	Email string `json:",omitempty"`
	Address string `json:",omitempty"`
	Remark string `json:",omitempty"`
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
func (c *DingTalkClient) DepartmentDetail(id int) (Department, error) {
    var data Department
    params := url.Values{}
    params.Add("id", fmt.Sprintf("%d", id))
    err :=c.httpRPC("department/get", params, nil, &data)
    return data, err
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

func (c *DingTalkClient) UserDetail(id string) (User, error) {
	var user User
	params := url.Values{}
	params.Add("userid", id)
	err :=c.httpRPC("user/get", params, nil, &user)
	return user, err
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
	return rep.DingtalkCorpExtAddResponse.UserID, nil
}

func (c *DingTalkClient) ExternalUserList(offset, size int) ([]ExternalUser, error) {
	var rep struct{
		TaobaoOAPIResponse
		DingtalkCorpExtListResponse struct{
			Result []ExternalUser
		} `json:"dingtalk_corp_ext_list_response"`
	}

	params := url.Values{}
	params.Add("size", strconv.Itoa(size))
	params.Add("offset", strconv.Itoa(offset))
	err := c.httpTaobaoRPC("dingtalk.corp.ext.list", params, &rep)
	if err != nil {
		return nil, err
	}
	return rep.DingtalkCorpExtListResponse.Result, nil
}

func (c *DingTalkClient) ExternalUserLabelGroups(offset, size int) ([]ExternalUserLabel, error) {
	var rep struct{
		TaobaoOAPIResponse
		DingtalkCorpExtListLabelGroupsResponse struct{
			Result []ExternalUserLabel
		} `json:"dingtalk_corp_ext_listlabelgroups_response"`
	}

	params := url.Values{}
	params.Add("size", strconv.Itoa(size))
	params.Add("offset", strconv.Itoa(offset))
	err := c.httpTaobaoRPC("dingtalk.corp.ext.listlabelgroups", params, &rep)
	if err != nil {
		return nil, err
	}
	return rep.DingtalkCorpExtListLabelGroupsResponse.Result, nil
}
