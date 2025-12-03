package handles

import (
    "path"
    "strconv"
    "strings"

    "github.com/OpenListTeam/OpenList/v4/internal/conf"
    "github.com/OpenListTeam/OpenList/v4/internal/errs"
    "github.com/OpenListTeam/OpenList/v4/internal/model"
    "github.com/OpenListTeam/OpenList/v4/internal/op"
    "github.com/OpenListTeam/OpenList/v4/internal/search"
    "github.com/OpenListTeam/OpenList/v4/pkg/utils"
    "github.com/OpenListTeam/OpenList/v4/server/common"
    "github.com/gin-gonic/gin"
    "github.com/pkg/errors"
)

type SearchReq struct {
	model.SearchReq
	Password string `json:"password"`
}

type SearchResp struct {
	model.SearchNode
	Type int `json:"type"`
}

func Search(c *gin.Context) {
	var (
		req SearchReq
		err error
	)
	if err = c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	user := c.Request.Context().Value(conf.UserKey).(*model.User)
	req.Parent, err = user.JoinPath(req.Parent)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if err := req.Validate(); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	nodes, total, err := search.Search(c, req.SearchReq)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
    var filteredNodes []model.SearchNode
    for _, node := range nodes {
        if !strings.HasPrefix(node.Parent, user.BasePath) {
            continue
        }
        meta, err := op.GetNearestMeta(node.Parent)
        if err != nil && !errors.Is(errors.Cause(err), errs.MetaNotFound) {
            continue
        }
        if !common.CanAccess(user, meta, path.Join(node.Parent, node.Name), req.Password) {
            continue
        }
        filteredNodes = append(filteredNodes, node)
    }
    if req.Distinct {
        seen := make(map[string]struct{})
        uniq := make([]model.SearchNode, 0, len(filteredNodes))
        for _, n := range filteredNodes {
            if n.IsDir {
                uniq = append(uniq, n)
                continue
            }
            key := strings.ToLower(n.Name) + "|" + strconv.FormatInt(n.Size, 10)
            if _, ok := seen[key]; ok {
                continue
            }
            seen[key] = struct{}{}
            uniq = append(uniq, n)
        }
        filteredNodes = uniq
    }
    respTotal := total
    if req.Distinct {
        respTotal = int64(len(filteredNodes))
    }
    common.SuccessResp(c, common.PageResp{
        Content: utils.MustSliceConvert(filteredNodes, nodeToSearchResp),
        Total:   respTotal,
    })
}

func nodeToSearchResp(node model.SearchNode) SearchResp {
	return SearchResp{
		SearchNode: node,
		Type:       utils.GetObjType(node.Name, node.IsDir),
	}
}
