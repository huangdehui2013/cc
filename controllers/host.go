package controllers

import (
	"encoding/json"
	"errors"
	"path"
	"strconv"
	"strings"

	"github.com/tealeg/xlsx"

	"github.com/shwinpiocess/cc/models"
)

// oprations for Host
type HostController struct {
	BaseController
}

// 导入主机
func (this *HostController) ImportPrivateHostByExcel() {
	out := make(map[string]interface{})
	f, h, err := this.GetFile("importPrivateHost")
	if f == nil {
		out["success"] = false
		out["message"] = "请先提供后缀名为xlsx的Excel文件再进行上传操作！"
		out["name"] = "importToCC"
		goto render
	}
	defer f.Close()
	if err != nil {
		out["success"] = false
		out["message"] = "主机导入失败！上传文件不合法！"
		out["name"] = "importToCC"
		goto render
	} else {
		if suffix := path.Ext(h.Filename); suffix != ".xlsx" {
			out["success"] = false
			out["message"] = "请提供后缀名为xlsx的Excel文件！"
			out["name"] = "importToCC"
			goto render
		}
		excelFileName := "static/upload/" + h.Filename
		this.SaveToFile("importPrivateHost", excelFileName)

		if xlFile, err := xlsx.OpenFile(excelFileName); err == nil {
			var hosts []*models.Host
			var ips string

			defApp, _ := models.GetDefAppByUserId(this.userId)
			for _, sheet := range xlFile.Sheets {
				for index, row := range sheet.Rows {
					if index > 0 && len(row.Cells) >= 14 {
						Model, err1 := row.Cells[0].String()
						Cpu, err2 := row.Cells[1].Int()
						Memory, err3 := row.Cells[2].Int()
						Hostname, err4 := row.Cells[3].String()
						InnerIp, err5 := row.Cells[4].String()
						InnerGate, err6 := row.Cells[5].String()
						InnerInterface, err7 := row.Cells[6].String()
						BgpIp, err8 := row.Cells[7].String()
						BgpGate, err9 := row.Cells[8].String()
						BgpInterface, err10 := row.Cells[9].String()
						OuterIp, err11 := row.Cells[10].String()
						OuterGate, err12 := row.Cells[11].String()
						OuterInterface, err13 := row.Cells[12].String()
						IloIp, err14 := row.Cells[13].String()
						if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil || err7 != nil || err8 != nil || err9 != nil || err10 != nil || err11 != nil || err12 != nil || err13 != nil || err14 != nil {
							out["success"] = false
							out["message"] = "主机导入失败！上传文件内容格式不正确！"
							out["name"] = "importToCC"
							goto render
						} else {
							if models.GetHostByInnerIp(InnerIp) {
								ips = ips + "<li>" + InnerIp + "</li>"
							}
							host := new(models.Host)
							host.Model = Model
							host.Cpu = Cpu
							host.Memory = Memory
							host.HostName = Hostname
							host.InnerIP = InnerIp
							host.InnerGate = InnerGate
							host.InnerInterface = InnerInterface
							host.BgpIP = BgpIp
							host.BgpGate = BgpGate
							host.BgpInterface = BgpInterface
							host.OuterIP = OuterIp
							host.OuterGate = OuterGate
							host.OuterInterface = OuterInterface
							host.IloIP = IloIp
							host.Source = 3
							host.ApplicationID = defApp["AppId"].(int)
							host.ApplicationName = defApp["AppName"].(string)
							host.SetID = defApp["SetId"].(int)
							host.SetName = defApp["SetName"].(string)
							host.ModuleID = defApp["ModuleId"].(int)
							host.ModuleName = defApp["ModuleName"].(string)
							hosts = append(hosts, host)
						}
					}
				}
			}

			if ips != "" {
				out["success"] = false
				out["message"] = `有内网IP在私有云中已经存在,请先修改这些IP的平台再做导入,具体如下：<ul class="">` + ips + "</ul>"
				out["name"] = "importToCC"
				goto render
			}

			if err := models.AddHost(hosts); err == nil {
				out["success"] = true
				out["message"] = "导入成功！"
				out["name"] = "importToCC"
			} else {
				out["success"] = false
				out["message"] = "主机导入数据库出现问题，请联系管理员！"
				out["name"] = "importToCC"
			}
		} else {
			out["success"] = false
			out["message"] = "主机导入失败！上传文件格式不正确！"
			out["name"] = "importToCC"
		}
	}

render:
	this.Data["result"] = out
	this.TplName = "host/upload.html"
}

// @Title Get
// @Description get Host by id
// @Param	id		path 	string	true		"The key for staticblock"
// @Success 200 {object} models.Host
// @Failure 403 :id is empty
// @router /:id [get]
func (this *HostController) Details() {
	//ApplicationID	5295
	// HostID	3668910
	id, _ := this.GetInt("HostID")
	_, err := models.GetHostById(id)
	if err != nil {
		this.TplName = "host/details.html"
	} else {
		this.TplName = "host/details.html"
	}
}

func (this *HostController) GetHost4QuickImport() {
	isDistributed, _ := this.GetBool("IsDistributed")
	source := this.GetString("Source")
	applicationId := this.GetString("ApplicationID")

	var fields []string
	var sortby []string
	var order []string
	var query = make(map[string]interface{})
	var limit int64 = 0
	var offset int64 = 0

	query["is_distributed"] = isDistributed
	query["source"] = source

	if isDistributed {
		query["application_id"] = applicationId
	}

	// fields: col1,col2,entity.col3
	if v := this.GetString("fields"); v != "" {
		fields = strings.Split(v, ",")
	}
	// limit: 10 (default is 10)
	if v, err := this.GetInt64("limit"); err == nil {
		limit = v
	}
	// offset: 0 (default is 0)
	if v, err := this.GetInt64("offset"); err == nil {
		offset = v
	}
	// sortby: col1,col2
	if v := this.GetString("sortby"); v != "" {
		sortby = strings.Split(v, ",")
	}
	// order: desc,asc
	if v := this.GetString("order"); v != "" {
		order = strings.Split(v, ",")
	}
	// query: k:v,k:v
	if v := this.GetString("query"); v != "" {
		for _, cond := range strings.Split(v, ",") {
			kv := strings.Split(cond, ":")
			if len(kv) != 2 {
				this.Data["json"] = errors.New("Error: invalid query key/value pair")
				this.ServeJSON()
				return
			}
			k, v := kv[0], kv[1]
			query[k] = v
		}
	}

	l, err := models.GetAllHost(query, fields, sortby, order, offset, limit)

	out := make(map[string]interface{})
	if err != nil {
		out["success"] = false
		this.jsonResult(out)
	} else {
		out["success"] = true
		out["data"] = l
		out["total"] = len(l)
		this.jsonResult(out)
	}
}

func (this *HostController) Put() {
	idStr := this.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	v := models.Host{HostID: id}
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &v); err == nil {
		if err := models.UpdateHostById(&v); err == nil {
			this.Data["json"] = "OK"
		} else {
			this.Data["json"] = err.Error()
		}
	} else {
		this.Data["json"] = err.Error()
	}
	this.ServeJSON()
}

// 删除主机
func (this *HostController) DelPrivateDefaultApplicationHost() {
	idStr := this.GetString("HostID")
	var ids []int
	for _, v := range strings.Split(idStr, ",") {
		if id, err := strconv.Atoi(v); err == nil {
			ids = append(ids, id)
		}
	}

	out := make(map[string]interface{})
	if _, err := models.DeleteHosts(ids); err == nil {
		out["success"] = true
		out["message"] = "删除成功!"
		this.jsonResult(out)
	} else {
		out["success"] = false
		out["errInfo"] = err.Error()
		this.jsonResult(out)
	}
}

// 分配主机
func (this *HostController) QuickDistribute() {
	// HostID:29,30
	// ApplicationID:4043
	// ToApplicationID:4041
	out := make(map[string]interface{})
	idStr := this.GetString("HostID")
	var ids []int
	for _, v := range strings.Split(idStr, ",") {
		if id, err := strconv.Atoi(v); err == nil {
			ids = append(ids, id)
		}
	}

	if toApplicationID, err := this.GetInt("ToApplicationID"); err != nil {
		out["success"] = false
		out["errInfo"] = err.Error()
		this.jsonResult(out)
	} else {
		if _, err := models.UpdateHostToApp(ids, toApplicationID); err != nil {
			out["success"] = false
			out["errInfo"] = err.Error()
			this.jsonResult(out)
		}

		out["success"] = true
		out["message"] = "分配成功"
		this.jsonResult(out)
	}
}

// 上交主机
func (this *HostController) ResHostModule() {
	// ApplicationID:4048
	// HostID:34
	out := make(map[string]interface{})
	idStr := this.GetString("HostID")
	var ids []int
	for _, v := range strings.Split(idStr, ",") {
		if id, err := strconv.Atoi(v); err == nil {
			ids = append(ids, id)
		}
	}

	if defApp, err := models.GetDefAppByUserId(this.userId); err != nil {
		out["success"] = false
		this.jsonResult(out)
	} else {
		if _, err := this.GetInt("ApplicationID"); err != nil {
			out["success"] = false
			out["errInfo"] = err.Error()
			out["message"] = err.Error()
			this.jsonResult(out)
		} else {
			//TODO 判断指定的业务是否存在
			if _, err := models.ResHostModule(ids, defApp["AppId"].(int)); err != nil {
				out["success"] = false
				out["errInfo"] = err
				out["message"] = err
				this.jsonResult(out)
			}

			out["success"] = true
			out["message"] = "上交成功"
			this.jsonResult(out)
		}
	}
}

// 主机管理
func (this *HostController) HostQuery() {
	info, options, _ := models.GetEmptyById(this.defaultApp.Id)
	this.Data["data"] = info
	this.Data["options"] = options
	if this.defaultApp.Level == 3 {
		this.TplName = "host/hostQuery_set.html"
	} else {
		this.TplName = "host/hostQuery_mod.html"
	}
}

// 快速分配
func (this *HostController) QuickImport() {
	this.TplName = "host/quickImport.html"
}

func (this *HostController) GetHostById() {
	out := make(map[string]interface{})
	//	var data []interface{}
	var fields []string
	var sortby []string
	var order []string
	var query = make(map[string]interface{})
	var limit int64 = 0
	var offset int64 = 0

	appId := this.GetString("ApplicationID")
	setId := this.GetString("SetID")
	modId := this.GetString("ModuleID")
	query["application_id"] = appId
	if setId != "" {
		query["set_id"] = setId
	}
	if modId != "" {
		query["module_id__in"] = strings.Split(modId, ",")
	}

	data, _ := models.GetAllHost(query, fields, sortby, order, offset, limit)
	out["data"] = data
	out["total"] = len(data)
	this.jsonResult(out)
}

func (this *HostController) GetTopoTree4view() {
	out := make(map[string]interface{})
	if appID, err := this.GetInt("ApplicationID"); err != nil {
		out["success"] = false
		out["message"] = "参数不合法！"
		this.jsonResult(out)
	} else {
		if data, _, err := models.GetEmptyById(appID); err != nil {
			out["success"] = false
			out["message"] = err
			this.jsonResult(out)
		} else {
			out = data
			this.jsonResult(out)
		}
	}
}

// 转移主机
func (this *HostController) ModHostModule() {
	// ApplicationID:4050
	// ModuleID:40
	// HostID:41,42,43,44,45,46,47,48,49
	var appID int
	var moduleID int
	var hostIds []int
	var id int
	var err error

	out := make(map[string]interface{})

	if appID, err = this.GetInt("ApplicationID"); err != nil {
		out["success"] = false
		out["message"] = "参数ApplicationID格式不正确"
		this.jsonResult(out)
	}

	if moduleID, err = this.GetInt("ModuleID"); err != nil {
		out["success"] = false
		out["message"] = "参数ModuleID格式不正确"
		this.jsonResult(out)
	}

	idStr := this.GetString("HostID")
	for _, v := range strings.Split(idStr, ",") {
		if id, err = strconv.Atoi(v); err != nil {
			out["success"] = false
			out["message"] = "参数HostID格式不正确"
			this.jsonResult(out)
		} else {
			hostIds = append(hostIds, id)
		}
	}

	if _, err = models.ModHostModule(appID, moduleID, hostIds); err != nil {
		out["success"] = false
		out["message"] = err.Error()
		this.jsonResult(out)
	} else {
		out["success"] = true
		out["message"] = "转移成功"
		this.jsonResult(out)
	}
}

// 移至空闲机/故障机
func (this *HostController) DelHostModule() {
	out := make(map[string]interface{})

	// ApplicationID:4050
	// HostID:41,42,45,46,47,48,49
	var appId int
	var hostIds []int
	var hostId int
	var moduleName string
	var err error

	if appId, err = this.GetInt("ApplicationID"); err != nil {
		out["success"] = false
		out["message"] = "参数ApplicationID格式不正确"
		this.jsonResult(out)
	}

	status := this.GetString("Status")
	if status == "1" {
		moduleName = "故障机"
	} else {
		moduleName = "空闲机"
	}

	for _, v := range strings.Split(this.GetString("HostID"), ",") {
		if hostId, err = strconv.Atoi(v); err != nil {
			out["success"] = false
			out["message"] = "参数HostID格式不正确"
			this.jsonResult(out)
		} else {
			hostIds = append(hostIds, hostId)
		}
	}

	if _, err = models.DelHostModule(appId, moduleName, hostIds); err != nil {
		out["success"] = false
		out["message"] = err.Error()
		this.jsonResult(out)
	} else {
		out["success"] = true
		out["message"] = "转移成功"
		this.jsonResult(out)
	}

}

// 查询主机
func (this *HostController) GetHostByCondition() {
	out := make(map[string]interface{})
	//  var data []interface{}
	var fields []string
	var sortby []string
	var order []string
	var query = make(map[string]interface{})
	var limit int64 = 0
	var offset int64 = 0

	appId := this.GetString("ApplicationID")
	modId := this.GetString("ModuleID")
	hostName := this.GetString("HostName")
	InnerIP := this.GetString("InnerIP")
	OuterIP := this.GetString("OuterIP")
	query["application_id"] = appId
	if hostName != "" {
		query["host_name__icontains"] = hostName
	}

	if InnerIP != "" {
		query["inner_ip__contains"] = InnerIP
	}
	if OuterIP != "" {
		query["outer_ip__contains"] = OuterIP
	}

	if modId != "" {
		query["module_id__in"] = strings.Split(modId, ",")
	}

	data, _ := models.GetAllHost(query, fields, sortby, order, offset, limit)
	out["data"] = data
	out["total"] = len(data)
	this.jsonResult(out)
}

// 更新主机信息
func (this *HostController) UpdateHostInfo() {
	out := make(map[string]interface{})
	HostID, _ := this.GetInt("HostID")
	ApplicationID, _ := this.GetInt("ApplicationID")
	HostName := this.GetString("HostName")
	InnerIP := this.GetString("InnerIP")
	OuterIP := this.GetString("OuterIP")
	h, err := models.GetHostById(HostID)
	if err != nil {
		out["success"] = false
		out["message"] = "主机不存在"
		this.jsonResult(out)
	}

	if ApplicationID != h.ApplicationID {
		out["success"] = false
		out["message"] = "没有操作权限"
		this.jsonResult(out)
	}

	if HostName != "" {
		h.HostName = HostName
	}

	if InnerIP != "" {
		h.InnerIP = InnerIP
	}

	if OuterIP != "" {
		h.OuterIP = OuterIP
	}

	err = models.UpdateHostById(h)
	if err != nil {
		out["success"] = false
		out["message"] = "更新失败"
		this.jsonResult(out)
	} else {
		out["success"] = true
		out["message"] = "更新成功"
		this.jsonResult(out)
	}
}
