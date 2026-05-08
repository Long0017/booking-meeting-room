package handler

import (
	"context"
	"net/http"

	"booking-api/internal/model"
)

func (h *Handler) listRooms(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(context.Background(),
		`SELECT id, name, capacity, COALESCE(location,''), is_active, created_at
		 FROM rooms WHERE is_active = true ORDER BY id`)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to fetch rooms")
		return
	}
	defer rows.Close()

	rooms := []model.Room{}
	for rows.Next() {
		var room model.Room
		if err := rows.Scan(&room.ID, &room.Name, &room.Capacity, &room.Location, &room.IsActive, &room.CreatedAt); err != nil {
			jsonError(w, http.StatusInternalServerError, "failed to scan room")
			return
		}
		rooms = append(rooms, room)
	}
	jsonOK(w, rooms)
}
