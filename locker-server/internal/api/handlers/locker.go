// ListLockers
// @Summary      사물함 목록 조회
// @Description  locker_info + locker_locations 조인 결과 반환
// @Tags         lockers
// @Security     BearerAuth
// @Produce      json
// @Success      200 {array}  struct{LockerID int `json:"locker_id"`; Owner *int `json:"owner,omitempty"`; Location string `json:"location"`}
// @Failure      401 {object} map[string]any
// @Router       /lockers [get]
package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Locker Response
type LockerResponse struct {
	LockerID   int    `json:"locker_id"`
	LocationID int    `json:"location_id"`
	Owner      string `json:"owner,omitempty"` // null 가능
}

// ListLockers: 사물함 목록 조회
// - locker_info + locker_locations 조인하여 위치명까지 함께 반환
// ListLockers godoc
// @Summary      사물함 목록 조회
// @Description  사용 가능한 사물함 목록을 반환합니다
// @Tags         lockers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} LockerResponse
// @Failure      401 {object} ErrorResponse
// @Router       /lockers [get]
func ListLockers(d Deps) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rows, err := d.DB.Query(c.Context(),
			`SELECT l.locker_id, l.owner, ll.name
			   FROM locker_info l
			   JOIN locker_locations ll ON ll.location_id = l.location_id
			   ORDER BY l.locker_id`)
		if err != nil {
			return fiber.ErrInternalServerError
		}
		defer rows.Close()

		var out []LockerResponse

		for rows.Next() {
			var it LockerResponse
			if err := rows.Scan(&it.LockerID, &it.Owner, &it.LocationID); err != nil {
				return fiber.ErrInternalServerError
			}
			out = append(out, it)
		}
		return c.JSON(out)
	}
}

// HoldLocker: 사물함 "선점"
// 1) Redis SETNX(key, student, TTL=2분) → 성공 시 첫 클릭 인정
// 2) DB에 locker_assignments(state='hold') 기록 (부분 유니크 인덱스로 중복 방지)
// - 실패 케이스: 이미 hold/confirmed가 존재 → 409
// HoldLocker godoc
// @Summary      사물함 선점
// @Description  특정 사물함을 선점합니다.
// @Tags         lockers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "사물함 ID"
// @Success      200 {object} LockerResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Router       /lockers/{id}/hold [post]
func HoldLocker(d Deps) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// URL 파라미터에서 locker id 추출
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return fiber.ErrBadRequest
		}

		// JWT 미들웨어에서 저장한 학번(sub)
		student, _ := c.Locals("student_id").(string)
		if student == "" {
			return fiber.ErrUnauthorized
		}

		// Redis 키: "locker:hold:{id}"
		key := "locker:hold:" + strconv.Itoa(id)

		// SETNX: 키가 없을 때만 set + TTL(2분). true=성공(첫 클릭), false=이미 누군가 보유중
		ok, err := d.RDB.SetNX(c.Context(), key, student, 2*time.Minute).Result()
		if err != nil {
			// Redis 장애 → 503(Service Unavailable)
			return fiber.ErrServiceUnavailable
		}
		if !ok {
			// 이미 다른 사람이 hold했거나, 본인이 선점했을 수도 있음 → 409
			return fiber.NewError(fiber.StatusConflict, "already held")
		}

		// DB 히스토리 기록 (hold)
		// * 유니크 인덱스가 마지막 안전망(한 locker/한 user당 활성 1건)
		_, err = d.DB.Exec(c.Context(),
			`INSERT INTO locker_assignments(locker_id, student_id, state, hold_expires_at)
			 VALUES ($1,$2,'hold', now() + interval '2 minutes')`,
			id, student)
		if err != nil {
			// DB에서 막히면 Redis 키를 지우는 게 깔끔(베스트 에포트)
			_, _ = d.RDB.Del(c.Context(), key).Result()
			return fiber.NewError(fiber.StatusConflict, "conflict")
		}

		// 성공 → 201
		return c.SendStatus(fiber.StatusCreated)
	}
}

// ConfirmLocker: "확정"
// - 내 hold가 유효(만료 전)해야 함
// - 트랜잭션으로 assignments를 confirmed로 바꾸고, locker_info.owner를 내 학번으로 설정
// ConfirmLocker godoc
// @Summary      사물함 확정
// @Description  선점한 사물함을 확정합니다.
// @Tags         lockers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "사물함 ID"
// @Success      200 {object} LockerResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Router       /lockers/{id}/confirm [post]
func ConfirmLocker(d Deps) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 파라미터 파싱
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return fiber.ErrBadRequest
		}
		student, _ := c.Locals("student_id").(string)
		if student == "" {
			return fiber.ErrUnauthorized
		}

		// 트랜잭션 시작
		tx, err := d.DB.Begin(c.Context())
		if err != nil {
			return fiber.ErrInternalServerError
		}
		defer tx.Rollback(c.Context())

		// 1) hold → confirmed 전환 (hold_expires_at 체크)
		ct, err := tx.Exec(c.Context(),
			`UPDATE locker_assignments
			   SET state='confirmed', confirmed_at=now()
			 WHERE locker_id=$1 AND student_id=$2
			   AND state='hold' AND (hold_expires_at IS NULL OR hold_expires_at > now())`,
			id, student)
		if err != nil || ct.RowsAffected() == 0 {
			// hold가 없거나 만료된 경우
			return fiber.NewError(fiber.StatusConflict, "hold expired or not found")
		}

		// 2) locker_info.owner를 내 학번으로 설정
		ct2, err := tx.Exec(c.Context(),
			`UPDATE locker_info SET owner=$1 WHERE locker_id=$2`, student, id)
		if err != nil || ct2.RowsAffected() == 0 {
			// 소유자 업데이트 실패 → 충돌 처리
			return fiber.ErrConflict
		}

		// 커밋
		if err := tx.Commit(c.Context()); err != nil {
			return fiber.ErrInternalServerError
		}

		// 성공 → 204 No Content
		return c.SendStatus(fiber.StatusNoContent)
	}
}

// ReleaseLocker: "해제"
// - confirmed 상태인 내 사물함을 취소하고, locker_info.owner=NULL
// ReleaseLocker godoc
// @Summary      사물함 해제
// @Description  선점한 사물함을 해제합니다.
// @Tags         lockers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "사물함 ID"
// @Success      200 {object} LockerResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Router       /lockers/{id}/release [post]
func ReleaseLocker(d Deps) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return fiber.ErrBadRequest
		}
		student, _ := c.Locals("student_id").(string)
		if student == "" {
			return fiber.ErrUnauthorized
		}

		// 트랜잭션
		tx, err := d.DB.Begin(c.Context())
		if err != nil {
			return fiber.ErrInternalServerError
		}
		defer tx.Rollback(c.Context())

		// 1) assignments: confirmed → cancelled
		_, err = tx.Exec(c.Context(),
			`UPDATE locker_assignments
			   SET state='cancelled', released_at=now()
			 WHERE locker_id=$1 AND student_id=$2 AND state='confirmed'`,
			id, student)
		if err != nil {
			return fiber.ErrInternalServerError
		}

		// 2) locker_info.owner=NULL (내가 소유자인 경우에만)
		_, err = tx.Exec(c.Context(),
			`UPDATE locker_info SET owner=NULL WHERE locker_id=$1 AND owner=$2`,
			id, student)
		if err != nil {
			return fiber.ErrInternalServerError
		}

		// 커밋
		if err := tx.Commit(c.Context()); err != nil {
			return fiber.ErrInternalServerError
		}

		// (옵션) 혹시 남아있을지 모르는 hold 키 제거(베스트 에포트)
		_, _ = d.RDB.Del(c.Context(), "locker:hold:"+strconv.Itoa(id)).Result()

		return c.SendStatus(fiber.StatusNoContent)
	}
}

/*
[중요 포인트 요약]
- Hold: Redis SETNX + TTL → 첫 클릭만 true. 이후 DB에 'hold' 기록.
- Confirm: DB 트랜잭션으로 유효 hold 확인 → confirmed 전환 + owner 설정.
- Release: confirmed를 cancelled로 → owner 해제.
- “최후의 안전망”: PostgreSQL 부분 유니크 인덱스(사물함당/사용자당 활성 1건).
*/
