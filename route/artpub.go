package route

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gokins-main/core/common"
	"github.com/gokins-main/core/utils"
	"github.com/gokins-main/gokins/comm"
	"github.com/gokins-main/gokins/model"
	"github.com/gokins-main/gokins/service"
	hbtp "github.com/mgr9525/HyperByte-Transfer-Protocol"
	"io"
	"net/url"
	"os"
	"path/filepath"
)

type ArtPublicController struct{}

func (ArtPublicController) GetPath() string {
	return "/api/art/pub"
}
func (c *ArtPublicController) Routes(g gin.IRoutes) {
	g.GET("/down/:id/:pth", c.down)
}
func (ArtPublicController) down(c *gin.Context) {
	id := c.Param("id")
	pth := c.Param("pth")
	tms := c.Query("times")
	random := c.Query("random")
	sign := c.Query("sign")

	artv := &model.TArtifactVersion{}
	ok := service.GetIdOrAid(artv, id)
	if !ok || artv.Deleted == 1 {
		c.String(404, "Not Found")
		return
	}
	arty := &model.TArtifactory{}
	ok = service.GetIdOrAid(arty, id)
	if !ok || arty.Deleted == 1 {
		c.String(404, "Not Found repo")
		return
	}

	lgusr, ok := service.CurrUserCache(c)
	if ok {
		perm := service.NewOrgPerm(lgusr, arty.OrgId)
		if !perm.CanDownload() {
			c.String(403, "No Permission")
			return
		}
	} else {
		signs := utils.Md5String(id + tms + random + comm.Cfg.Server.DevDownToken)
		if sign != signs {
			c.String(403, "No Permission")
			return
		}
	}

	fls := filepath.Join(comm.WorkPath, common.PathArtifacts, artv.Id, pth)
	stat, err := os.Stat(fls)
	if err != nil {
		c.String(404, "Not Found File")
		return
	}
	var nms string
	var contsz int64
	var rdr io.Reader
	if stat.IsDir() {
		zipth := filepath.Join(comm.WorkPath, common.PathTmp, utils.NewXid())
		err = utils.Zip(fls, zipth)
		if err != nil {
			c.String(511, "Zip err1:"+err.Error())
			return
		}
		defer os.RemoveAll(zipth)
		stat, err = os.Stat(zipth)
		if err != nil {
			c.String(511, "Zip err2:"+err.Error())
			return
		}
		fls = zipth
		nms = stat.Name() + ".zip"
	} else {
		nms = stat.Name()
	}

	fl, err := os.Open(fls)
	if err != nil {
		c.String(404, "Not Found File")
		return
	}
	defer fl.Close()
	rdr = fl
	contsz = stat.Size()

	c.Header("Connection", "Keep-Alive")
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Cache-Control", "max-age=360000000")
	c.Header("Content-Length", fmt.Sprintf("%d", contsz))
	c.Header("Content-Disposition", fmt.Sprintf(`attachment;filename="%s"`, url.QueryEscape(nms)))
	c.Status(200)
	bts := make([]byte, 10240)
	for !hbtp.EndContext(c) {
		n, err := rdr.Read(bts)
		if n > 0 {
			c.Writer.Write(bts[:n])
		}
		if err != nil {
			break
		}
	}
}