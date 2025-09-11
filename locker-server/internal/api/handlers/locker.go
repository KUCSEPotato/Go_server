package handlers

import (
	"os"
	"strconv"
	"time"

	"github.com/KUCSEPotato/locker-server/internal/scheduler"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

var lockerApplicationStart time.Time
var lockerApplicationStartErr error
var lockerApplicationEnd time.Time
var lockerApplicationEndErr error

func init() {
	startStr := os.Getenv("LOCKER_APPLICATION_START")
	if startStr != "" {
		lockerApplicationStart, lockerApplicationStartErr = time.Parse(time.RFC3339, startStr)
	}

	endStr := os.Getenv("LOCKER_APPLICATION_END")
	if endStr != "" {
		lockerApplicationEnd, lockerApplicationEndErr = time.Parse(time.RFC3339, endStr)
	}
}

// Locker Response
type LockerResponse struct {
	LockerID   int     `json:"locker_id"`
	LocationID string  `json:"location_id"`
	Owner      *string `json:"owner,omitempty"` // null 가능
	// RemainingCount int     `json:"remaining_count"` // 사용 가능한 사물함 수
}

// Hold Success Response
type HoldSuccessResponse struct {
	Message   string         `json:"message" example:"locker held successfully"`
	Locker    LockerResponse `json:"locker"`
	ExpiresIn string         `json:"expires_in" example:"5 minutes"`
}

// Hold Fallback Response
type HoldFallbackResponse struct {
	Message   string `json:"message" example:"locker held successfully"`
	LockerID  int    `json:"locker_id" example:"101"`
	ExpiresIn string `json:"expires_in" example:"5 minutes"`
}

// List Lockers Response
type ListLockersResponse struct {
	Lockers        []LockerResponse `json:"lockers"`
	AvailableCount int              `json:"available_count" example:"45"`
}

// Simple Success Response
type SimpleSuccessResponse struct {
	Message string `json:"message" example:"operation completed successfully"`
}

// My Locker Response
type MyLockerResponse struct {
	Locker *LockerResponse `json:"locker"`
}

// ListLockers: 사물함 목록 조회
// - locker_info + locker_locations 조인하여 위치명까지 함께 반환
// ListLockers godoc
// @Summary      사물함 목록 조회
// @Description  모든 사물함 목록과 사용 가능한 사물함 수를 반환합니다
// @Tags         lockers
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer {access_token}" default(Bearer )
// @Success      200 {object} ListLockersResponse "사물함 목록과 사용 가능 수"
// @Failure      401 {object} ErrorResponse "인증 필요"
// @Failure      500 {object} ErrorResponse "서버 오류"
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

		// 단일 응답에 사용 가능한 사물함 수 포함
		var availableCount int
		err = d.DB.QueryRow(c.Context(),
			`SELECT COUNT(*) FROM locker_info WHERE owner IS NULL`).Scan(&availableCount)
		if err != nil {
			return fiber.ErrInternalServerError
		}
		return c.JSON(ListLockersResponse{
			Lockers:        out,
			AvailableCount: availableCount,
		})
	}
}

// HoldLocker: 사물함 "선점"
// 1) Redis SETNX(key, student, TTL=5분) → 성공 시 첫 클릭 인정
// 2) DB에 locker_assignments(state='hold') 기록 (부분 유니크 인덱스로 중복 방지)
// - 실패 케이스: 이미 hold/confirmed가 존재 → 409
// HoldLocker godoc
// @Summary      사물함 선점
// @Description  특정 사물함을 선점합니다 (5분간 예약). Redis와 DB를 통해 동시성 제어를 하며, 성공 시 사물함 정보를 반환합니다. 신청 기간 외에는 접근이 불가능합니다.
// @Tags         lockers
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer {access_token}" default(Bearer )
// @Param        id path int true "사물함 ID (1-999 범위)" minimum(1) maximum(999) example(101)
// @Success      201 {object} HoldSuccessResponse "선점 성공 - 사물함 정보 포함"
// @Failure      400 {object} ErrorResponse "잘못된 요청 - 유효하지 않은 사물함 ID"
// @Failure      401 {object} ErrorResponse "인증 필요 - JWT 토큰이 없거나 유효하지 않음"
// @Failure      403 {object} ErrorResponse "신청 기간 외 - 신청 시작 전이거나 마감 후"
// @Failure      409 {object} ErrorResponse "이미 선점됨 - 다른 사용자가 이미 선점했거나 본인이 이미 선점한 상태"
// @Failure      503 {object} ErrorResponse "서비스 일시 불가 - Redis 서버 장애"
// @Router       /lockers/{id}/hold [post]
func HoldLocker(d Deps) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 신청 시작 시간 체크
		now := time.Now()
		if lockerApplicationStartErr == nil && now.Before(lockerApplicationStart) {
			return fiber.NewError(fiber.StatusForbidden, "아직 신청 기간이 아닙니다. 신청 시작: "+lockerApplicationStart.Format("2006-01-02 15:04:05"))
		}

		// 신청 마감 시간 체크
		if lockerApplicationEndErr == nil && now.After(lockerApplicationEnd) {
			return fiber.NewError(fiber.StatusForbidden, "신청 기간이 마감되었습니다. 신청 마감: "+lockerApplicationEnd.Format("2006-01-02 15:04:05"))
		}

		// URL 파라미터에서 locker id 추출
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return fiber.ErrBadRequest
		}

		// 해당 locker의 만료된 hold를 먼저 정리
		scheduler.CheckAndCleanupExpiredHold(d.DB, d.RDB, id)

		// JWT 미들웨어에서 저장한 학번(sub)
		student, _ := c.Locals("student_id").(string)
		if student == "" {
			return fiber.ErrUnauthorized
		}

		// Redis 키: "locker:hold:{id}"
		key := "locker:hold:" + strconv.Itoa(id)

		// SETNX: 키가 없을 때만 set + TTL(5분). true=성공(첫 클릭), false=이미 누군가 보유중
		// [250908] 1분으로 변경 테스트
		ok, err := d.RDB.SetNX(c.Context(), key, student, 1*time.Minute).Result()
		if err != nil {
			// Redis 장애 → 503(Service Unavailable)
			return fiber.ErrServiceUnavailable
		}
		if !ok {
			// 이미 다른 사람이 hold했거나, 본인이 선점했을 수도 있음 → 409
			return fiber.NewError(fiber.StatusConflict, "Locker already held by someone")
		}

		// DB 히스토리 기록 (hold)
		// * 유니크 인덱스가 마지막 안전망(한 locker/한 user당 활성 1건)
		_, err = d.DB.Exec(c.Context(),
			`INSERT INTO locker_assignments(locker_id, student_id, state, hold_expires_at)
			 VALUES ($1,$2,'hold', now() + interval '5 minutes')`,
			id, student)
		if err != nil {
			// DB에서 막히면 Redis 키를 삭제(베스트 에포트)
			_, _ = d.RDB.Del(c.Context(), key).Result()
			return fiber.NewError(fiber.StatusConflict, "Locker hold failed on DB. Deleting Redis key.")
		}

		// 성공 시 사물함 정보도 함께 반환
		var lockerInfo LockerResponse
		err = d.DB.QueryRow(c.Context(),
			`SELECT l.locker_id, l.owner, ll.name
			 FROM locker_info l
			 JOIN locker_locations ll ON ll.location_id = l.location_id
			 WHERE l.locker_id = $1`,
			id).Scan(&lockerInfo.LockerID, &lockerInfo.Owner, &lockerInfo.LocationID)
		if err != nil {
			// 정보 조회 실패해도 hold는 성공했으므로 기본 정보만 반환
			return c.Status(fiber.StatusCreated).JSON(HoldFallbackResponse{
				Message:   "locker held successfully",
				LockerID:  id,
				ExpiresIn: "5 minutes",
			})
		}

		// 성공 → 201 + 사물함 정보
		return c.Status(fiber.StatusCreated).JSON(HoldSuccessResponse{
			Message:   "locker held successfully",
			Locker:    lockerInfo,
			ExpiresIn: "5 minutes",
		})
	}
}

// ConfirmLocker: "확정"
// - 내 hold가 유효(만료 전)해야 함
// - 트랜잭션으로 assignments를 confirmed로 바꾸고, locker_info.owner를 내 학번으로 설정
// ConfirmLocker godoc
// @Summary      사물함 확정
// @Description  선점한 사물함을 확정합니다 (실제 소유권 획득). hold 상태에서 confirmed 상태로 전환되며, 사물함의 소유자로 등록됩니다.
// @Tags         lockers
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer {access_token}" default(Bearer )
// @Param        id path int true "사물함 ID (선점한 사물함)" minimum(1) maximum(999) example(101)
// @Success      200 {object} SimpleSuccessResponse "확정 완료"
// @Failure      400 {object} ErrorResponse "잘못된 요청 - 유효하지 않은 사물함 ID"
// @Failure      401 {object} ErrorResponse "인증 필요 - JWT 토큰이 없거나 유효하지 않음"
// @Failure      409 {object} ErrorResponse "선점이 만료되었거나 없음 - hold 상태가 아니거나 5분이 경과함"
// @Failure      500 {object} ErrorResponse "서버 오류 - 데이터베이스 트랜잭션 실패"
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
			`UPDATE locker_info SET owner=$1 WHERE locker_id=$2 AND owner IS NULL`, student, id)
		if err != nil || ct2.RowsAffected() == 0 {
			// 소유자 업데이트 실패 → 충돌 처리
			return fiber.ErrConflict
		}

		// 커밋
		if err := tx.Commit(c.Context()); err != nil {
			return fiber.ErrInternalServerError
		}

		// 성공 → 200
		return c.JSON(SimpleSuccessResponse{
			Message: "locker confirmed successfully",
		})
	}
}

/*
// DirectConfirmLocker: 바로 확정 (Hold 단계 생략)
func DirectConfirmLocker(d Deps) fiber.Handler {
    return func(c *fiber.Ctx) error {
        id, err := strconv.Atoi(c.Params("id"))
        if err != nil {
            return fiber.ErrBadRequest
        }

        student, _ := c.Locals("student_id").(string)
        if student == "" {
            return fiber.ErrUnauthorized
        }

        // 트랜잭션으로 한 번에 처리
        tx, err := d.DB.Begin(c.Context())
        if err != nil {
            return fiber.ErrInternalServerError
        }
        defer tx.Rollback(c.Context())

        // 1) 사물함이 비어있는지 확인하고 바로 소유자 설정
        ct, err := tx.Exec(c.Context(),
            `UPDATE locker_info SET owner=$1 WHERE locker_id=$2 AND owner IS NULL`,
            student, id)
        if err != nil || ct.RowsAffected() == 0 {
            return fiber.NewError(fiber.StatusConflict, "locker already taken")
        }

        // 2) 히스토리 기록 (confirmed 상태로 바로)
        _, err = tx.Exec(c.Context(),
            `INSERT INTO locker_assignments(locker_id, student_id, state, confirmed_at)
             VALUES ($1,$2,'confirmed', now())`,
            id, student)
        if err != nil {
            return fiber.ErrInternalServerError
        }

        if err := tx.Commit(c.Context()); err != nil {
            return fiber.ErrInternalServerError
        }

        return c.JSON(fiber.Map{"message": "locker confirmed successfully"})
    }
}
*/

// ReleaseLocker: "해제"
// - confirmed 상태인 내 사물함을 취소하고, locker_info.owner=NULL
// ReleaseLocker godoc
// @Summary      사물함 해제
// @Description  확정된 사물함을 해제합니다 (소유권 포기). confirmed 상태에서 cancelled 상태로 전환되며, 사물함이 다시 사용 가능해집니다.
// @Tags         lockers
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer {access_token}" default(Bearer )
// @Param        id path int true "사물함 ID (소유한 사물함)" minimum(1) maximum(999) example(101)
// @Success      200 {object} SimpleSuccessResponse "해제 완료"
// @Failure      400 {object} ErrorResponse "잘못된 요청 - 유효하지 않은 사물함 ID"
// @Failure      401 {object} ErrorResponse "인증 필요 - JWT 토큰이 없거나 유효하지 않음"
// @Failure      404 {object} ErrorResponse "사물함을 찾을 수 없음 - 소유하지 않은 사물함"
// @Failure      500 {object} ErrorResponse "서버 오류 - 데이터베이스 트랜잭션 실패"
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
		ct, err := tx.Exec(c.Context(),
			`UPDATE locker_assignments
			   SET state='cancelled', released_at=now()
			 WHERE locker_id=$1 AND student_id=$2 AND state='confirmed'`,
			id, student)
		if err != nil {
			return fiber.ErrInternalServerError
		}
		if ct.RowsAffected() == 0 {
			return fiber.NewError(fiber.StatusNotFound, "No confirmed locker found to release")
		}

		// 2) locker_info.owner=NULL (내가 소유자인 경우에만)
		ct, err = tx.Exec(c.Context(),
			`UPDATE locker_info SET owner=NULL WHERE locker_id=$1 AND owner=$2`,
			id, student)
		if err != nil {
			return fiber.ErrInternalServerError
		}
		if ct.RowsAffected() == 0 {
			return fiber.NewError(fiber.StatusNotFound, "No locker ownership found to release")
		}

		// 3) Remove any lingering hold assignments
		_, err = tx.Exec(c.Context(),
			`DELETE FROM locker_assignments WHERE locker_id=$1 AND student_id=$2 AND state='hold'`,
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

		return c.JSON(SimpleSuccessResponse{
			Message: "locker released successfully",
		})
	}
}

// ReleaseHold: "hold 해제"
// - hold 상태인 사물함 예약을 취소합니다.
// ReleaseHold godoc
// @Summary      사물함 hold 해제
// @Description  hold 상태인 사물함 예약을 취소합니다. 사물함이 다시 사용 가능해집니다.
// @Tags         lockers
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer {access_token}" default(Bearer )
// @Param        id path int true "사물함 ID (hold 상태의 사물함)" minimum(1) maximum(999) example(101)
// @Success      200 {object} SimpleSuccessResponse "hold 해제 완료"
// @Failure      400 {object} ErrorResponse "잘못된 요청 - 유효하지 않은 사물함 ID"
// @Failure      401 {object} ErrorResponse "인증 필요 - JWT 토큰이 없거나 유효하지 않음"
// @Failure      404 {object} ErrorResponse "사물함을 찾을 수 없음 - hold 상태가 아님"
// @Failure      500 {object} ErrorResponse "서버 오류 - 데이터베이스 트랜잭션 실패"
// @Router       /lockers/{id}/release-hold [post]
func ReleaseHold(d Deps) fiber.Handler {
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

		// hold 상태 해제
		ct, err := tx.Exec(c.Context(),
			`DELETE FROM locker_assignments WHERE locker_id=$1 AND student_id=$2 AND state='hold'`,
			id, student)
		if err != nil {
			return fiber.ErrInternalServerError
		}
		if ct.RowsAffected() == 0 {
			return fiber.NewError(fiber.StatusNotFound, "No hold found to release")
		}

		// 커밋
		if err := tx.Commit(c.Context()); err != nil {
			return fiber.ErrInternalServerError
		}

		// Redis 키 제거 (베스트 에포트)
		_, _ = d.RDB.Del(c.Context(), "locker:hold:"+strconv.Itoa(id)).Result()

		return c.JSON(SimpleSuccessResponse{
			Message: "hold released successfully",
		})
	}
}

// GetMyLocker: 현재 로그인한 유저가 소유한 사물함 조회
// GetMyLocker godoc
// @Summary      내 사물함 조회
// @Description  현재 로그인한 사용자가 소유한 사물함 정보를 반환합니다. 소유한 사물함이 없으면 locker가 null로 반환됩니다.
// @Tags         lockers
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer {access_token}" default(Bearer )
// @Success      200 {object} MyLockerResponse "내 사물함 정보 (없으면 locker가 null)"
// @Failure      401 {object} ErrorResponse "인증 필요 - JWT 토큰이 없거나 유효하지 않음"
// @Failure      500 {object} ErrorResponse "서버 오류 - 데이터베이스 조회 실패"
// @Router       /lockers/me [get]
func GetMyLocker(d Deps) fiber.Handler {
	return func(c *fiber.Ctx) error {
		student, _ := c.Locals("student_id").(string)
		if student == "" {
			return fiber.ErrUnauthorized
		}

		var it LockerResponse
		err := d.DB.QueryRow(c.Context(),
			`SELECT l.locker_id, l.owner, ll.name
               FROM locker_info l
               JOIN locker_locations ll ON ll.location_id = l.location_id
              WHERE l.owner = $1`,
			student,
		).Scan(&it.LockerID, &it.Owner, &it.LocationID)

		if err != nil {
			if err == pgx.ErrNoRows {
				return c.JSON(MyLockerResponse{Locker: nil})
			}
			return fiber.ErrInternalServerError
		}

		return c.JSON(MyLockerResponse{Locker: &it})
	}
}

/*
[중요 포인트 요약]
- Hold: Redis SETNX + TTL → 첫 클릭만 true. 이후 DB에 'hold' 기록.
- Confirm: DB 트랜잭션으로 유효 hold 확인 → confirmed 전환 + owner 설정.
- Release: confirmed를 cancelled로 → owner 해제.
- “최후의 안전망”: PostgreSQL 부분 유니크 인덱스(사물함당/사용자당 활성 1건).
*/
