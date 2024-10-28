package metrics

import (
	"encoding/json"
	"net/http"

	"ClamGuardian/internal/position"
)

// FileStatusHandler 处理文件状态请求
func FileStatusHandler(pm *position.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		positions := pm.GetAllPositions()

		type fileStatus struct {
			Filename string  `json:"filename"`
			Position int64   `json:"position"`
			Size     int64   `json:"size"`
			Progress float64 `json:"progress"`
		}

		status := make([]fileStatus, 0, len(positions))
		for _, pos := range positions {
			progress := float64(0)
			if pos.FileSize > 0 {
				progress = float64(pos.Position) / float64(pos.FileSize)
			}
			status = append(status, fileStatus{
				Filename: pos.Filename,
				Position: pos.Position,
				Size:     pos.FileSize,
				Progress: progress,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	}
}
