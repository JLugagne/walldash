package converters

import (
	"github.com/JLugagne/walldash/internal/dashboard/domain"
	pkgdashboard "github.com/JLugagne/walldash/pkg/dashboard"
)

// ToDomainAction converts a public ActionRequest to a domain ActionCommand.
func ToDomainAction(req pkgdashboard.ActionRequest) domain.ActionCommand {
	return domain.ActionCommand{
		EntityID: req.EntityID,
		Action:   req.Action,
	}
}

// ToPublicAction converts a domain ActionCommand to a public ActionResponse.
func ToPublicAction(cmd domain.ActionCommand) pkgdashboard.ActionResponse {
	return pkgdashboard.ActionResponse{
		Status:   "success",
		EntityID: cmd.EntityID,
		Action:   cmd.Action,
	}
}
