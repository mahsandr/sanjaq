package post

import "go.uber.org/zap"

// Log errors that do not need to stop the service and are at the warning level
func (p *Handler) checkError(errOrigin string, err error) {
	if err != nil {
		p.logger.Warn(errOrigin+" error",
			zap.Error(err))
	}
}
