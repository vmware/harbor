package security

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	robotCtx "github.com/goharbor/harbor/src/common/security/robot"
	"github.com/goharbor/harbor/src/common/utils"
	robot_ctl "github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/pkg/robot/model"
	"net/http"
)

type robot2 struct{}

func (r *robot2) Generate(req *http.Request) security.Context {
	log := log.G(req.Context())
	name, secret, ok := req.BasicAuth()
	if !ok {
		return nil
	}
	if !strings.HasPrefix(name, config.RobotPrefix(req.Context())) {
		return nil
	}
	// The robot name can be used as the unique identifier to locate robot as it contains the project name.
	robots, err := robot_ctl.Ctl.List(req.Context(), q.New(q.KeyWords{
		"name": strings.TrimPrefix(name, config.RobotPrefix(req.Context())),
	}), &robot_ctl.Option{
		WithPermission: true,
	})
	if err != nil {
		log.Errorf("failed to list robots: %v", err)
		return nil
	}
	if len(robots) == 0 {
		return nil
	}

	robot := robots[0]
	if utils.Encrypt(secret, robot.Salt, utils.SHA256) != robot.Secret {
		log.Errorf("failed to authenticate robot account: %s", name)
		return nil
	}
	if robot.Disabled {
		log.Errorf("failed to authenticate disabled robot account: %s", name)
		return nil
	}
	now := time.Now().Unix()
	if robot.ExpiresAt != -1 && robot.ExpiresAt <= now {
		log.Errorf("the robot account is expired: %s", name)
		return nil
	}

	var accesses []*types.Policy
	for _, p := range robot.Permissions {
		for _, a := range p.Access {
			var resource string
			if a.Resource == rbac.ResourceProject && p.Kind == "project" {
				resource = p.Scope
			} else {
				resource = fmt.Sprintf("%s/%s", p.Scope, a.Resource)
			}

			accesses = append(accesses, &types.Policy{
				Action:   a.Action,
				Effect:   a.Effect,
				Resource: types.Resource(resource),
			})
		}
	}

	modelRobot := &model.Robot{
		Name: name,
	}
	log.Infof("a robot2 security context generated for request %s %s", req.Method, req.URL.Path)
	return robotCtx.NewSecurityContext(modelRobot, robot.Level == robot_ctl.LEVELSYSTEM, accesses)
}
