package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

// User 유저 정보
// @Description users 테이블의 한 레코드(민감정보 제외)
type User struct {
	StudentID   string `json:"student_id" example:"2025320000"`
	Name        string `json:"name"        example:"홍길동"`
	PhoneNumber string `json:"phone_number" example:"01012345678"`
}

// ListUsersResponse 전체 유저 목록 응답
type ListUsersResponse struct {
	Count int    `json:"count" example:"42"`
	Users []User `json:"users"`
}

// GetAllUsersHandler godoc
// @Summary      모든 유저 조회
// @Description  users 테이블의 전체 유저를 조회하고, 총 개수도 함께 반환합니다.
// @Tags         users
// @Produce      json
// @Success      200 {object} handlers.ListUsersResponse
// @Failure      500 {object} map[string]string
// @Router       /users/GetAllUsers [get]
func GetAllUsersHandler(db *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		query := `
			SELECT student_id, name, phone_number
			FROM users
			ORDER BY student_id ASC
		`

		rows, err := db.Query(context.Background(), query)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to query users")
		}
		defer rows.Close()

		var users []User
		for rows.Next() {
			var u User
			if err := rows.Scan(&u.StudentID, &u.Name, &u.PhoneNumber); err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "failed to scan user row")
			}
			users = append(users, u)
		}

		var total int
		err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM users").Scan(&total)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to count users")
		}

		return c.JSON(ListUsersResponse{
			Count: total,
			Users: users,
		})
	}
}
