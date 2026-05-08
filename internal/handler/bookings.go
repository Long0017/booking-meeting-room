package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"booking-api/internal/model"
)

var bangkokLoc, _ = time.LoadLocation("Asia/Bangkok")

func (h *Handler) listBookings(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	from := q.Get("from")
	to := q.Get("to")

	query := `
		SELECT b.id, b.room_id, r.name, b.employee_id, b.department, b.title,
		       b.start_time, b.end_time, b.attendees_count, COALESCE(b.phone,''), COALESCE(b.notes,''), b.status, b.created_at
		FROM bookings b JOIN rooms r ON r.id = b.room_id
		WHERE b.status = 'confirmed'`

	args := []any{}

	if from != "" && to != "" {
		fromT, errFrom := time.ParseInLocation("2006-01-02", from, bangkokLoc)
		toT, errTo := time.ParseInLocation("2006-01-02", to, bangkokLoc)
		if errFrom == nil && errTo == nil {
			toT = toT.Add(24 * time.Hour)
			query += ` AND b.start_time < $1 AND b.end_time > $2`
			args = append(args, toT, fromT)
		}
	}

	query += ` ORDER BY b.start_time`

	rows, err := h.db.Query(context.Background(), query, args...)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to fetch bookings")
		return
	}
	defer rows.Close()

	bookings := []model.Booking{}
	for rows.Next() {
		var b model.Booking
		if err := rows.Scan(&b.ID, &b.RoomID, &b.RoomName, &b.EmployeeID, &b.Department,
			&b.Title, &b.StartTime, &b.EndTime, &b.AttendeesCount, &b.Phone, &b.Notes, &b.Status, &b.CreatedAt); err != nil {
			jsonError(w, http.StatusInternalServerError, "failed to scan booking")
			return
		}
		bookings = append(bookings, b)
	}
	jsonOK(w, bookings)
}

func (h *Handler) createBooking(w http.ResponseWriter, r *http.Request) {
	var req model.CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.EmployeeID) == "" {
		jsonError(w, http.StatusBadRequest, "ກະລຸນາໃສ່ລະຫັດພະນັກງານ")
		return
	}
	if strings.TrimSpace(req.Department) == "" {
		jsonError(w, http.StatusBadRequest, "ກະລຸນາລະບຸພາກສ່ວນ")
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		jsonError(w, http.StatusBadRequest, "ກະລຸນາໃສ່ຊື່ກອງປະຊຸມ")
		return
	}
	if strings.TrimSpace(req.CancelCode) == "" {
		jsonError(w, http.StatusBadRequest, "ກະລຸນາໃສ່ລະຫັດຢືນຢັນ")
		return
	}
	if len(strings.TrimSpace(req.CancelCode)) < 4 {
		jsonError(w, http.StatusBadRequest, "ລະຫັດຢືນຢັນຕ້ອງມີຢ່າງໜ້ອຍ 4 ຕົວ")
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "ຮູບແບບ start_time ບໍ່ຖືກຕ້ອງ (RFC3339)")
		return
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "ຮູບແບບ end_time ບໍ່ຖືກຕ້ອງ (RFC3339)")
		return
	}

	if !endTime.After(startTime) {
		jsonError(w, http.StatusBadRequest, "ເວລາສິ້ນສຸດຕ້ອງຫຼາຍກວ່າເວລາເລີ່ມຕົ້ນ")
		return
	}
	if req.AttendeesCount < 1 {
		req.AttendeesCount = 1
	}
	if req.RoomID == 0 {
		req.RoomID = 1
	}

	var id int
	err = h.db.QueryRow(context.Background(),
		`INSERT INTO bookings (room_id, employee_id, department, title, start_time, end_time, attendees_count, phone, notes, cancel_code)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`,
		req.RoomID, req.EmployeeID, req.Department, req.Title,
		startTime, endTime, req.AttendeesCount, strings.TrimSpace(req.Phone), req.Notes, strings.TrimSpace(req.CancelCode),
	).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23P01" {
			jsonError(w, http.StatusConflict, "ໄລຍະເວລານີ້ຖືກຈອງແລ້ວ ກະລຸນາເລືອກໄລຍະເວລາອື່ນ")
			return
		}
		jsonError(w, http.StatusInternalServerError, "ບໍ່ສາມາດບັນທຶກການຈອງໄດ້: "+err.Error())
		return
	}

	var b model.Booking
	h.db.QueryRow(context.Background(),
		`SELECT b.id, b.room_id, r.name, b.employee_id, b.department, b.title,
		        b.start_time, b.end_time, b.attendees_count, COALESCE(b.phone,''), COALESCE(b.notes,''), b.status, b.created_at
		 FROM bookings b JOIN rooms r ON r.id = b.room_id WHERE b.id = $1`, id,
	).Scan(&b.ID, &b.RoomID, &b.RoomName, &b.EmployeeID, &b.Department,
		&b.Title, &b.StartTime, &b.EndTime, &b.AttendeesCount, &b.Phone, &b.Notes, &b.Status, &b.CreatedAt)

	jsonCreated(w, b)
}

func (h *Handler) cancelBooking(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, http.StatusBadRequest, "invalid booking id")
		return
	}

	var req model.CancelBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(req.CancelCode) == "" {
		jsonError(w, http.StatusBadRequest, "ກະລຸນາໃສ່ລະຫັດຢືນຢັນ")
		return
	}

	var storedCode, status string
	err = h.db.QueryRow(context.Background(),
		`SELECT cancel_code, status FROM bookings WHERE id = $1`, id,
	).Scan(&storedCode, &status)

	if errors.Is(err, pgx.ErrNoRows) {
		jsonError(w, http.StatusNotFound, "ບໍ່ພົບຂໍ້ມູນການຈອງ")
		return
	}
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to find booking")
		return
	}
	if storedCode != strings.TrimSpace(req.CancelCode) {
		jsonError(w, http.StatusForbidden, "ລະຫັດຢືນຢັນບໍ່ຖືກຕ້ອງ ບໍ່ສາມາດຍົກເລີກໄດ້")
		return
	}
	if status == "cancelled" {
		jsonError(w, http.StatusBadRequest, "ການຈອງນີ້ຖືກຍົກເລີກແລ້ວ")
		return
	}

	if _, err = h.db.Exec(context.Background(),
		`UPDATE bookings SET status = 'cancelled' WHERE id = $1`, id); err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to cancel booking")
		return
	}
	jsonOK(w, map[string]string{"message": "ຍົກເລີກການຈອງສໍາເລັດ"})
}
