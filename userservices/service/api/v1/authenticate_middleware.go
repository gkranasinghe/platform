package v1

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	commonService "github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/userservices/client"
	"github.com/tidepool-org/platform/userservices/service"
)

func Authenticate(handler service.HandlerFunc) service.HandlerFunc {
	return func(context service.Context) {
		userSessionToken := context.Request().Header.Get(client.TidepoolUserSessionTokenHeaderName)
		if userSessionToken == "" {
			context.RespondWithError(commonService.ErrorAuthenticationTokenMissing())
			return
		}

		requestUserID, err := context.UserServicesClient().ValidateUserSession(context, userSessionToken)
		if err != nil {
			if client.IsUnauthorizedError(err) {
				context.RespondWithError(commonService.ErrorUnauthenticated())
			} else {
				context.RespondWithInternalServerFailure("Unable to validate user session", err, userSessionToken)
			}
			return
		}

		context.SetRequestUserID(requestUserID)

		handler(context)
	}
}
