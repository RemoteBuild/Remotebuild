package handlers

import (
	"net/http"
	"os/exec"

	"github.com/JojiiOfficial/Remotebuild/models"
	log "github.com/sirupsen/logrus"
)

// Clear local ccache
func clearCcache(handlerData HandlerData, w http.ResponseWriter, r *http.Request) {
	// Clear ccache
	out, err := exec.Command("ccache", "-c").Output()

	if err != nil {
		log.Error("Error cleaning ccache: ", err)
		sendResponse(w, models.ResponseError, err.Error(), nil)
		return
	}

	sendResponse(w, models.ResponseSuccess, string(out), nil)
}

// Get ccache stats
func ccacheStats(handlerData HandlerData, w http.ResponseWriter, r *http.Request) {
	// Clear ccache
	out, err := exec.Command("sh", "-c", "ccache -s | grep -v config").Output()

	if err != nil {
		log.Error("Error querying ccache: ", err)
		sendResponse(w, models.ResponseError, err.Error(), nil)
		return
	}

	sendResponse(w, models.ResponseSuccess, "", models.StringResponse{
		String: string(out),
	})
}
