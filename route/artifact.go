package route

import (
	"github.com/gin-gonic/gin"
	"github.com/gokins-main/core/utils"
	"github.com/gokins-main/gokins/bean"
	"github.com/gokins-main/gokins/comm"
	"github.com/gokins-main/gokins/model"
	"github.com/gokins-main/gokins/models"
	"github.com/gokins-main/gokins/service"
	"github.com/gokins-main/gokins/util"
	hbtp "github.com/mgr9525/HyperByte-Transfer-Protocol"
	"net/http"
	"strings"
	"time"
)

type ArtifactController struct{}

func (ArtifactController) GetPath() string {
	return "/api/art"
}
func (c *ArtifactController) Routes(g gin.IRoutes) {
	g.Use(service.MidUserCheck)
	g.POST("/org-list", util.GinReqParseJson(c.orgList))
	g.POST("/edit", util.GinReqParseJson(c.edit))
	g.POST("/rm", util.GinReqParseJson(c.rm))
}
func (ArtifactController) orgList(c *gin.Context, m *hbtp.Map) {
	orgId := m.GetString("orgId")
	q := m.GetString("q")
	pg, _ := m.GetInt("page")
	if orgId == "" {
		c.String(500, "param err")
		return
	}
	lgusr := service.GetMidLgUser(c)
	perm := service.NewOrgPerm(lgusr, orgId)
	if perm.Org() == nil || perm.Org().Deleted == 1 {
		c.String(404, "not found org")
		return
	}
	if !perm.CanRead() {
		c.String(405, "No Auth")
		return
	}
	ls := make([]*models.TArtifactory, 0)
	var err error
	var page *bean.Page
	if comm.IsMySQL {
		gen := &bean.PageGen{
			CountCols: "art.aid",
			FindCols:  "art.*",
		}
		gen.SQL = `
			select {{select}} from t_artifactory art 
			where art.deleted != 1 and art.org_id=?
		    `
		gen.Args = append(gen.Args, perm.Org().Id)
		if q != "" {
			gen.SQL += "\nAND art.name like ? "
			gen.Args = append(gen.Args, "%"+q+"%")
		}
		gen.SQL += "\nORDER BY art.aid DESC"
		page, err = comm.FindPages(gen, &ls, pg, 20)
		if err != nil {
			c.String(500, "db err:"+err.Error())
			return
		}
	}
	for _, v := range ls {
		usr, ok := service.GetUser(v.Uid)
		if ok {
			v.Nick = usr.Nick
			v.Avat = usr.Avatar
		}
		e := &model.TArtifactPackage{}
		v.Artln, _ = comm.Db.Where("repo_id=?", v.Id).Count(e)
	}
	c.JSON(http.StatusOK, page)
}
func (ArtifactController) edit(c *gin.Context, m *hbtp.Map) {
	orgId := m.GetString("orgId")
	id := m.GetString("id")
	name := strings.TrimSpace(m.GetString("name"))
	desc := strings.TrimSpace(m.GetString("desc"))
	disabled := m.GetBool("disabled")
	if name == "" {
		c.String(500, "param err")
		return
	}
	lgusr := service.GetMidLgUser(c)
	perm := service.NewOrgPerm(lgusr, orgId)
	if perm.Org() == nil || perm.Org().Deleted == 1 {
		c.String(404, "not found org")
		return
	}
	if !perm.IsOrgAdmin() {
		c.String(405, "No Permission")
		return
	}
	var err error
	ne := &model.TArtifactory{}
	isup := service.GetIdOrAid(id, ne)
	ne.Name = name
	ne.Desc = desc
	if disabled {
		ne.Disabled = 1
	} else {
		ne.Disabled = 0
	}
	ne.Updated = time.Now()
	if isup {
		if ne.OrgId != perm.Org().Id {
			c.String(405, "No Permission")
			return
		}
		_, err = comm.Db.Cols("name", "desc", "disabled", "updated").Where("id=?", ne.Id).Update(ne)
	} else {
		ne.Id = utils.NewXid()
		ne.Uid = lgusr.Id
		ne.OrgId = perm.Org().Id
		ne.Created = time.Now()

		ln := 0
		ne.Identifier = strings.ToLower(utils.RandomString(8))
		for !hbtp.EndContext(c) {
			ln++
			n, _ := comm.Db.Where("identifier=?", ne.Identifier).Count(ne)
			if n <= 0 {
				break
			}
			i := 8
			if ln >= 9 {
				i = 11
			} else if ln >= 6 {
				i = 10
			} else if ln >= 3 {
				i = 9
			}
			ne.Identifier = strings.ToLower(utils.RandomString(i))
		}
		_, err = comm.Db.InsertOne(ne)
	}
	if err != nil {
		c.String(500, "db err:"+err.Error())
		return
	}
	c.String(200, ne.Id)
}
func (ArtifactController) rm(c *gin.Context, m *hbtp.Map) {
	id := m.GetString("id")
	art := &model.TArtifactory{}
	ok := service.GetIdOrAid(id, art)
	if !ok {
		c.String(404, "Not Found")
		return
	}
	lgusr := service.GetMidLgUser(c)
	perm := service.NewOrgPerm(lgusr, art.OrgId)
	if perm.Org() == nil || perm.Org().Deleted == 1 {
		c.String(404, "not found org")
		return
	}
	if !perm.IsOrgAdmin() {
		c.String(405, "No Permission")
		return
	}
	art.Deleted = 1
	art.DeletedTime = time.Now()
	art.Updated = time.Now()
	_, err := comm.Db.Cols("deleted", "deleted_time", "updated").Where("id=?", art.Id).Update(art)
	if err != nil {
		c.String(500, "db err:"+err.Error())
		return
	}
	c.String(200, art.Id)
}